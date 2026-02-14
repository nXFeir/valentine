package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Config struct {
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	RefreshToken    string
	Sender          string
	Recipients      []string
	Subject         string
	Date            string
	GifURL          string
	CorsOrigin      string
	EnableOAuthFlow bool
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}

	var gmailSvc *gmail.Service
	if cfg.RefreshToken != "" {
		tokenSource := oauthCfg.TokenSource(ctx, &oauth2.Token{RefreshToken: cfg.RefreshToken})
		httpClient := oauth2.NewClient(ctx, tokenSource)
		gmailSvc, err = gmail.NewService(ctx, option.WithHTTPClient(httpClient))
		if err != nil {
			log.Fatalf("gmail service: %v", err)
		}
	} else {
		log.Println("warning: GMAIL_REFRESH_TOKEN is empty, /api/yes will fail")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/yes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w, cfg)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		writeCORSHeaders(w, cfg)
		if gmailSvc == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "email service not configured"})
			return
		}

		if err := sendValentineEmail(ctx, gmailSvc, cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	})

	if cfg.EnableOAuthFlow {
		mux.HandleFunc("/oauth/start", func(w http.ResponseWriter, r *http.Request) {
			url := oauthCfg.AuthCodeURL("valentine", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
			http.Redirect(w, r, url, http.StatusFound)
		})

		mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing code"})
				return
			}

			token, err := oauthCfg.Exchange(ctx, code)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{"refresh_token": token.RefreshToken})
		})
	}

	port := envOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() (Config, error) {
	cfg := Config{
		ClientID:        os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret:    os.Getenv("GMAIL_CLIENT_SECRET"),
		RedirectURL:     os.Getenv("GMAIL_REDIRECT_URL"),
		RefreshToken:    os.Getenv("GMAIL_REFRESH_TOKEN"),
		Sender:          os.Getenv("GMAIL_SENDER"),
		Subject:         envOrDefault("EMAIL_SUBJECT", "Valentine Date"),
		Date:            envOrDefault("EMAIL_DATE", "March 14, 2026"),
		GifURL:          envOrDefault("EMAIL_GIF_URL", "https://media.giphy.com/media/3oEjI4sFlp73fvEYgw/giphy.gif"),
		CorsOrigin:      envOrDefault("CORS_ORIGIN", "*"),
		EnableOAuthFlow: strings.EqualFold(os.Getenv("ENABLE_OAUTH_FLOW"), "true"),
	}

	recipients := strings.TrimSpace(os.Getenv("EMAIL_RECIPIENTS"))
	if recipients != "" {
		cfg.Recipients = splitAndTrim(recipients)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
		return cfg, errors.New("missing GMAIL_CLIENT_ID, GMAIL_CLIENT_SECRET, or GMAIL_REDIRECT_URL")
	}

	if cfg.Sender == "" {
		return cfg, errors.New("missing GMAIL_SENDER")
	}

	if len(cfg.Recipients) == 0 {
		return cfg, errors.New("missing EMAIL_RECIPIENTS")
	}

	return cfg, nil
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

func writeCORSHeaders(w http.ResponseWriter, cfg Config) {
	w.Header().Set("Access-Control-Allow-Origin", cfg.CorsOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func sendValentineEmail(ctx context.Context, svc *gmail.Service, cfg Config) error {
	bodyLines := []string{
		"We are officially booked for our valentine date!",
		fmt.Sprintf("Date: %s", cfg.Date),
		"",
		"Cute gif:",
		cfg.GifURL,
	}

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", cfg.Sender),
		fmt.Sprintf("To: %s", strings.Join(cfg.Recipients, ", ")),
		fmt.Sprintf("Subject: %s", cfg.Subject),
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		strings.Join(bodyLines, "\n"),
	}, "\r\n")

	raw := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(message))
	_, err := svc.Users.Messages.Send("me", &gmail.Message{Raw: raw}).Do()
	return err
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
