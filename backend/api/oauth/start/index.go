package handler

import (
	"net/http"
	"os"
	"strings"

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

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		state = "valentine"
	}

	url := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusFound)
}
