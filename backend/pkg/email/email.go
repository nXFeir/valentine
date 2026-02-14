package email

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

//go:embed spongebob-patric.gif
var embeddedGif []byte

type Config struct {
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	RefreshToken     string
	Sender           string
	Recipients       []string
	Subject          string
	Date             string
	GifURL           string
	EventTitle       string
	EventDescription string
	EventDate        string
	EventAllDay      bool
	EventTimeZone    string
	CorsOrigin       string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ClientID:         os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret:     os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:      os.Getenv("GMAIL_REDIRECT_URL"),
		RefreshToken:     os.Getenv("GMAIL_REFRESH_TOKEN"),
		Sender:           os.Getenv("GMAIL_SENDER"),
		Subject:          envOrDefault("EMAIL_SUBJECT", "Valentine Date"),
		Date:             envOrDefault("EMAIL_DATE", "March 14, 2026"),
		GifURL:           envOrDefault("EMAIL_GIF_URL", "https://media.giphy.com/media/3oEjI4sFlp73fvEYgw/giphy.gif"),
		EventTitle:       envOrDefault("EVENT_TITLE", "Valentine Date"),
		EventDescription: envOrDefault("EVENT_DESCRIPTION", "Can't wait to celebrate together."),
		EventDate:        envOrDefault("EVENT_DATE", "2026-03-14"),
		EventAllDay:      envOrDefault("EVENT_ALL_DAY", "true") == "true",
		EventTimeZone:    envOrDefault("EVENT_TIMEZONE", "Asia/Kuala_Lumpur"),
		CorsOrigin:       envOrDefault("CORS_ORIGIN", "*"),
	}

	recipients := strings.TrimSpace(os.Getenv("EMAIL_RECIPIENTS"))
	if recipients != "" {
		cfg.Recipients = splitAndTrim(recipients)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
		return cfg, errors.New("missing GMAIL_CLIENT_ID, GMAIL_CLIENT_SECRET, or GMAIL_REDIRECT_URL")
	}

	if cfg.RefreshToken == "" {
		return cfg, errors.New("missing GMAIL_REFRESH_TOKEN")
	}

	if cfg.Sender == "" {
		return cfg, errors.New("missing GMAIL_SENDER")
	}

	if len(cfg.Recipients) == 0 {
		return cfg, errors.New("missing EMAIL_RECIPIENTS")
	}

	return cfg, nil
}

func Send(ctx context.Context, cfg Config) error {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}

	tokenSource := oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: cfg.RefreshToken})
	httpClient := oauth2.NewClient(ctx, tokenSource)
	gmailSvc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return err
	}

	message, err := BuildMessage(cfg, time.Now().UTC(), fmt.Sprintf("valentine-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}

	raw := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(message))
	_, err = gmailSvc.Users.Messages.Send("me", &gmail.Message{Raw: raw}).Do()
	return err
}

func BuildMessage(cfg Config, now time.Time, uid string) (string, error) {
	gifBytes := embeddedGif
	if len(gifBytes) == 0 {
		return "", fmt.Errorf("embedded gif is empty")
	}

	calendarInvite, err := buildCalendarInvite(cfg, now, uid)
	if err != nil {
		return "", err
	}

	boundaryMixed := "mixed_" + uid
	boundaryAlt := "alt_" + uid
	boundaryRelated := "rel_" + uid
	contentID := "valentine-gif"

	htmlBody := fmt.Sprintf(
		"<html><body><p>We are officially booked for our valentine date!</p><p><strong>Date:</strong> %s</p><img src=\"cid:%s\" alt=\"Valentine gif\" /></body></html>",
		cfg.Date,
		contentID,
	)

	plainBody := strings.Join([]string{
		"We are officially booked for our valentine date!",
		fmt.Sprintf("Date: %s", cfg.Date),
		"",
		"Cute gif:",
		cfg.GifURL,
	}, "\n")

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("From: %s\r\n", cfg.Sender))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(cfg.Recipients, ", ")))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", cfg.Subject))
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundaryMixed))

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
	builder.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundaryAlt))

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlt))
	builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	builder.WriteString(plainBody)
	builder.WriteString("\r\n\r\n")

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlt))
	builder.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=\"%s\"\r\n\r\n", boundaryRelated))

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryRelated))
	builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	builder.WriteString(htmlBody)
	builder.WriteString("\r\n\r\n")

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryRelated))
	builder.WriteString("Content-Type: image/gif\r\n")
	builder.WriteString("Content-Transfer-Encoding: base64\r\n")
	builder.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", contentID))
	builder.WriteString("Content-Disposition: inline; filename=\"spongebob-patric.gif\"\r\n\r\n")
	builder.WriteString(chunkBase64(gifBytes))
	builder.WriteString("\r\n\r\n")

	builder.WriteString(fmt.Sprintf("--%s--\r\n\r\n", boundaryRelated))

	builder.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlt))
	builder.WriteString("Content-Type: text/calendar; charset=\"UTF-8\"; method=REQUEST; name=\"invite.ics\"\r\n")
	builder.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	builder.WriteString("Content-Disposition: attachment; filename=\"invite.ics\"\r\n")
	builder.WriteString("Content-Class: urn:content-classes:calendarmessage\r\n\r\n")
	builder.WriteString(calendarInvite)
	builder.WriteString("\r\n\r\n")

	builder.WriteString(fmt.Sprintf("--%s--\r\n\r\n", boundaryAlt))

	builder.WriteString(fmt.Sprintf("--%s--\r\n", boundaryMixed))
	return builder.String(), nil
}

func buildCalendarInvite(cfg Config, now time.Time, uid string) (string, error) {
	eventDate, err := time.Parse("2006-01-02", cfg.EventDate)
	if err != nil {
		return "", fmt.Errorf("invalid EVENT_DATE: %w", err)
	}

	dtStamp := now.UTC().Format("20060102T150405Z")
	start := eventDate.Format("20060102")
	end := eventDate.AddDate(0, 0, 1).Format("20060102")

	var builder strings.Builder
	builder.WriteString("BEGIN:VCALENDAR\r\n")
	builder.WriteString("PRODID:-//Valentine//EN\r\n")
	builder.WriteString("VERSION:2.0\r\n")
	builder.WriteString("CALSCALE:GREGORIAN\r\n")
	builder.WriteString("METHOD:REQUEST\r\n")
	builder.WriteString(fmt.Sprintf("X-WR-TIMEZONE:%s\r\n", cfg.EventTimeZone))
	builder.WriteString("BEGIN:VEVENT\r\n")
	builder.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
	builder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", dtStamp))
	builder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(cfg.EventTitle)))
	builder.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(cfg.EventDescription)))
	builder.WriteString(fmt.Sprintf("ORGANIZER:mailto:%s\r\n", cfg.Sender))

	for _, recipient := range cfg.Recipients {
		builder.WriteString(fmt.Sprintf("ATTENDEE;CN=%s;ROLE=REQ-PARTICIPANT;RSVP=TRUE:mailto:%s\r\n", recipient, recipient))
	}

	if cfg.EventAllDay {
		builder.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", start))
		builder.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", end))
	}

	builder.WriteString("END:VEVENT\r\n")
	builder.WriteString("END:VCALENDAR\r\n")
	return builder.String(), nil
}

func escapeICS(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		";", "\\;",
		",", "\\,",
		"\n", "\\n",
		"\r", "",
	)
	return replacer.Replace(value)
}

func chunkBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLen = 76
	var builder strings.Builder
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		builder.WriteString(encoded[i:end])
		builder.WriteString("\r\n")
	}
	return builder.String()
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			trimmed = append(trimmed, value)
		}
	}
	return trimmed
}
