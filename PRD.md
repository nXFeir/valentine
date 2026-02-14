# Product Requirements Document (PRD)

## Overview
A single-page, romantic website that asks the user to be my valentine. The page features a heartfelt prompt with a puppy-eyes GIF, a clear "Yes" button, and a playful "No" button that jumps to new positions when approached. On "Yes," the GIF switches to a celebratory reaction.

## Goals
- Deliver a romantic, memorable proposal experience.
- Make the "No" option playfully difficult to click without hiding it entirely.
- Provide a clear, delightful confirmation state after "Yes."
- Work smoothly on both desktop and mobile devices.

## Non-Goals
- No authentication, user accounts, or data collection.
- No backend or persistence.
- No multi-page navigation.

## Target Audience
- A single recipient (my girlfriend), viewing on mobile or desktop.

## Tone and Visual Style
- Romantic, sweet, and playful.
- Clean layout with a single focal area: text + GIF + buttons.

## User Flow
1. User lands on the page and sees the romantic prompt and puppy-eyes GIF.
2. Two buttons are shown: "Yes" and "No."
3. If the user approaches "No" (hover on desktop, tap attempt on mobile), the "No" button jumps away.
4. If the user clicks "Yes," the GIF switches to a celebratory GIF and a short confirmation message appears.

## Content Requirements
- Primary prompt (romantic tone), example:
  - "Will you be my valentine?"
- Secondary supportive line (optional), example:
  - "I really hope you say yes."
- "Yes" button label: "Yes"
- "No" button label: "No"
- Confirmation message after "Yes," example:
  - "Yay! You made me the happiest person."

## Interaction Requirements
### "No" Button Behavior
- The "No" button jumps to a new random position when:
  - Desktop: pointer hovers over it.
  - Mobile: user taps it (or attempts to tap it).
- The button is allowed to move partially or briefly off-screen.
- If it moves off-screen, it must return to a visible position within 1 to 1.5 seconds.
- The button should remain visible most of the time (no extended hiding).
- Minimum move distance: at least 80px from its previous position.
- Cooldown: at least 250ms between jumps to avoid jitter.
- The "No" button is not keyboard-focusable and does not respond to keyboard focus.

### "Yes" Button Behavior
- Clicking "Yes" switches the GIF to a celebratory GIF.
- The confirmation message appears immediately.
- The "Yes" button is not keyboard-focusable.

## GIF and Asset Requirements
- Default GIF: puppy with puppy eyes (beggy eyes).
- Success GIF: celebratory, happy reaction.
- GIFs should be embedded as local assets or safe external links.
- Provide graceful fallback if a GIF fails to load (e.g., alt text).

## Accessibility Note
- Keyboard navigation is intentionally excluded for both buttons for playful effect.

## Responsive Design Requirements
- Layout adapts to small screens with a vertical stack (text, GIF, buttons).
- Buttons remain large enough for tapping on mobile.
- "No" button movement respects screen bounds and reappears promptly.

## Acceptance Criteria
- The initial view shows the romantic prompt, puppy-eyes GIF, and both buttons.
- On desktop hover over "No," the button jumps to a new position.
- On mobile tap attempt of "No," the button jumps away.
- "No" can briefly move off-screen but reappears within 1 to 1.5 seconds.
- The "No" button moves at least 80px each jump and respects a 250ms cooldown.
- Clicking "Yes" swaps the GIF to the celebratory GIF and shows confirmation text.
- The experience works on common desktop and mobile viewport sizes.
- Neither button is reachable via keyboard navigation.

