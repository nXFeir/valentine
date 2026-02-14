package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GMAIL_CLIENT_ID")
	clientSecret := os.Getenv("GMAIL_CLIENT_SECRET")
	redirectURL := os.Getenv("GMAIL_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		http.Error(w, "Missing Gmail OAuth configuration", http.StatusInternalServerError)
		return
	}

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{gmail.GmailSendScope},
		Endpoint:     google.Endpoint,
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"refresh_token": token.RefreshToken,
	})
}
