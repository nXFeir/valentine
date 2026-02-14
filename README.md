# Valentine Proposal Site

A single-page, romantic proposal website with a playful, jumpy "No" button and a celebratory GIF swap on "Yes."

## Run
Open `index.html` directly in a browser, or serve it locally.

```powershell
# Option 1: open directly
start .\index.html

# Option 2: run a simple local server (Python 3)
python -m http.server 5500
```

Then visit `http://localhost:5500`.

## Backend (Go + Gmail API)
This backend sends the email automatically when the user clicks "Yes." It exposes `POST /api/yes`.

### Required environment variables
- `GMAIL_CLIENT_ID`
- `GMAIL_CLIENT_SECRET`
- `GMAIL_REDIRECT_URL`
- `GMAIL_REFRESH_TOKEN`
- `GMAIL_SENDER`
- `EMAIL_RECIPIENTS` (comma-separated)

Optional:
- `EMAIL_SUBJECT` (default: "Valentine Date")
- `EMAIL_DATE` (default: "March 14, 2026")
- `EMAIL_GIF_URL`
- `CORS_ORIGIN` (default: `*`)
- `ENABLE_OAUTH_FLOW` (set to `true` to enable `/oauth/start` + `/oauth/callback`)

### Frontend API base URL
If your backend is hosted on a different domain, set this before loading `script.js`:

```html
<script>
  window.VALENTINE_API_BASE = "https://your-backend.example.com";
</script>
```

### Run backend locally
```powershell
cd .\backend
# Install deps and run
Go mod download
Go run .
```

### Vercel deployment note
If you point Vercel to the `backend` folder, routes are available at:
- `https://<your-domain>/api/health`
- `https://<your-domain>/api/yes`

The `backend/main.go` file is for local server usage; Vercel uses the functions under `backend/api/`.

## Customize the GIFs
Update the GIF URLs or paths in `script.js`:

```javascript
const GIFS = {
  ask: "<your puppy-eyes gif>",
  celebrate: "<your celebratory gif>",
};
```

You can use external URLs or local files (e.g., `assets/puppy-eyes.gif`).

## Notes
- The "No" button jumps on hover (desktop) and tap (mobile), with brief off-screen allowance.
- Keyboard navigation is intentionally disabled for both buttons.
