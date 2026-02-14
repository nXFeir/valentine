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

