const GIFS = {
  ask: "puppy-dog-eyes-please.gif",
  celebrate: "despicable-me-minions.gif",
};

const API_BASE_URL = window.VALENTINE_API_BASE || "";

const NO_TEXTS = [
  "No",
  "Nope",
  "Try again",
  "Nice try",
  "Still no",
  "Click Yes",
  "BLEHH",
  "Almost!",
  "HEHE",
];

const MIN_DISTANCE = 80;
const COOLDOWN_MS = 250;
const OFFSCREEN_RETURN_MIN_MS = 1000;
const OFFSCREEN_RETURN_MAX_MS = 1500;

const mainGif = document.getElementById("main-gif");
const confirmMessage = document.getElementById("confirm-message");
const yesButton = document.getElementById("yes-button");
const noButton = document.getElementById("no-button");

let lastMoveTime = 0;
let lastPosition = null;
let returnTimer = null;
let lastNoText = null;

mainGif.src = GIFS.ask;

const randomInRange = (min, max) => Math.random() * (max - min) + min;

const pickNoText = () => {
  if (NO_TEXTS.length === 1) {
    lastNoText = NO_TEXTS[0];
    return lastNoText;
  }

  let nextText = NO_TEXTS[0];
  for (let i = 0; i < 8; i += 1) {
    nextText = NO_TEXTS[Math.floor(Math.random() * NO_TEXTS.length)];
    if (nextText !== lastNoText) {
      break;
    }
  }

  lastNoText = nextText;
  return nextText;
};

const updateNoLabel = () => {
  noButton.textContent = pickNoText();
};

const sendYesEmail = async () => {
  const response = await fetch(`${API_BASE_URL}/api/yes`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({}),
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Failed to send email");
  }
};

const getViewport = () => ({
  width: document.documentElement.clientWidth,
  height: document.documentElement.clientHeight,
});

const getButtonSize = () => {
  const rect = noButton.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
};

const getBounds = (allowOffscreen) => {
  const { width, height } = getViewport();
  const { width: btnWidth, height: btnHeight } = getButtonSize();
  const pad = allowOffscreen ? Math.round(Math.min(btnWidth, btnHeight) * 0.6) : 0;

  return {
    minX: -pad,
    maxX: width - btnWidth + pad,
    minY: -pad,
    maxY: height - btnHeight + pad,
  };
};

const isOffscreen = (pos) => {
  const { width, height } = getViewport();
  const { width: btnWidth, height: btnHeight } = getButtonSize();

  return (
    pos.x < 0 ||
    pos.y < 0 ||
    pos.x + btnWidth > width ||
    pos.y + btnHeight > height
  );
};

const pickPosition = (allowOffscreen) => {
  const bounds = getBounds(allowOffscreen);
  let candidate = null;

  for (let i = 0; i < 12; i += 1) {
    const x = randomInRange(bounds.minX, bounds.maxX);
    const y = randomInRange(bounds.minY, bounds.maxY);
    if (!lastPosition) {
      candidate = { x, y };
      break;
    }

    const distance = Math.hypot(x - lastPosition.x, y - lastPosition.y);
    if (distance >= MIN_DISTANCE) {
      candidate = { x, y };
      break;
    }

    candidate = { x, y };
  }

  return candidate;
};

const setPosition = (pos) => {
  noButton.style.left = `${pos.x}px`;
  noButton.style.top = `${pos.y}px`;
  lastPosition = pos;
};

const moveNoButton = ({ allowOffscreen, force = false } = {}) => {
  const now = Date.now();
  if (!force && now - lastMoveTime < COOLDOWN_MS) {
    return;
  }

  lastMoveTime = now;
  if (returnTimer) {
    clearTimeout(returnTimer);
    returnTimer = null;
  }

  const nextPosition = pickPosition(allowOffscreen);
  setPosition(nextPosition);
  updateNoLabel();

  if (allowOffscreen && isOffscreen(nextPosition)) {
    // Force a visible return after a short delay when we jump off-screen.
    const delay = randomInRange(OFFSCREEN_RETURN_MIN_MS, OFFSCREEN_RETURN_MAX_MS);
    returnTimer = setTimeout(() => {
      moveNoButton({ allowOffscreen: false, force: true });
    }, delay);
  }
};

const ensureVisible = () => {
  const bounds = getBounds(false);
  const x = Math.min(Math.max(lastPosition?.x ?? bounds.minX, bounds.minX), bounds.maxX);
  const y = Math.min(Math.max(lastPosition?.y ?? bounds.minY, bounds.minY), bounds.maxY);
  setPosition({ x, y });
};

const placeInitialNoButton = () => {
  const { width, height } = getViewport();
  const { width: btnWidth, height: btnHeight } = getButtonSize();
  const x = (width - btnWidth) / 2 + 120;
  const y = Math.min(height - btnHeight - 40, height * 0.6);
  setPosition({ x, y });
  updateNoLabel();
  ensureVisible();
};

noButton.addEventListener("pointerenter", (event) => {
  if (event.pointerType === "mouse") {
    moveNoButton({ allowOffscreen: true });
  }
});

noButton.addEventListener("pointerdown", (event) => {
  moveNoButton({ allowOffscreen: true, force: true });
  event.preventDefault();
  event.stopPropagation();
});

noButton.addEventListener("click", (event) => {
  event.preventDefault();
});

yesButton.addEventListener("click", async () => {
  mainGif.src = GIFS.celebrate;
  mainGif.alt = "Celebratory reaction";
  confirmMessage.hidden = false;
  confirmMessage.textContent = "Sending our invitation...";

  try {
    await sendYesEmail();
    confirmMessage.textContent = "Yay! The invitation is on its way.";
    yesButton.disabled = true;
  } catch (error) {
    confirmMessage.textContent = "Hmm, I could not send the email. Please try again.";
  }
});

window.addEventListener("resize", () => {
  if (lastPosition) {
    ensureVisible();
  }
});

window.addEventListener("load", () => {
  mainGif.src = GIFS.ask;
  placeInitialNoButton();
});
