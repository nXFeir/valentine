package email

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	RefreshToken string
	Sender       string
	Recipients   []string
	Subject      string
	Date         string
	GifURL       string
	CorsOrigin   string
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GMAIL_REDIRECT_URL"),
		RefreshToken: os.Getenv("GMAIL_REFRESH_TOKEN"),
		Sender:       os.Getenv("GMAIL_SENDER"),
		Subject:      envOrDefault("EMAIL_SUBJECT", "Valentine Date"),
		Date:         envOrDefault("EMAIL_DATE", "March 14, 2026"),
		GifURL:       envOrDefault("EMAIL_GIF_URL", "https://media.giphy.com/media/3oEjI4sFlp73fvEYgw/giphy.gif"),
		CorsOrigin:   envOrDefault("CORS_ORIGIN", "*"),
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

	message := buildMessage(cfg)
	raw := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(message))
	_, err = gmailSvc.Users.Messages.Send("me", &gmail.Message{Raw: raw}).Do()
	return err
}

func buildMessage(cfg Config) string {
	bodyLines := []string{
		"We are officially booked for our valentine date!",
		fmt.Sprintf("Date: %s", cfg.Date),
		"",
		"Cute gif:",
		cfg.GifURL,
	}

	headers := []string{
		fmt.Sprintf("From: %s", cfg.Sender),
		fmt.Sprintf("To: %s", strings.Join(cfg.Recipients, ", ")),
		fmt.Sprintf("Subject: %s", cfg.Subject),
		"Content-Type: text/plain; charset=\"UTF-8\"",
	}

	return strings.Join(append(headers, "", strings.Join(bodyLines, "\n")), "\r\n")
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
