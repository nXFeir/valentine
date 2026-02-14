package email

import (
	"strings"
	"testing"
	"time"
)

func TestBuildMessageIncludesCalendarAndGif(t *testing.T) {
	cfg := Config{
		Sender:           "sender@example.com",
		Recipients:       []string{"recipient@example.com"},
		Subject:          "Valentine Date",
		Date:             "March 14, 2026",
		GifURL:           "https://example.com/gif",
		EventTitle:       "Valentine Date",
		EventDescription: "Can't wait",
		EventDate:        "2026-03-14",
		EventAllDay:      true,
		EventTimeZone:    "Asia/Kuala_Lumpur",
	}

	msg, err := BuildMessage(cfg, time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC), "test-uid")
	if err != nil {
		t.Fatalf("BuildMessage error: %v", err)
	}

	if !strings.Contains(msg, "Content-Type: text/calendar") {
		t.Fatalf("expected calendar invite")
	}

	if !strings.Contains(msg, "Content-Type: image/gif") {
		t.Fatalf("expected inline gif")
	}

	if !strings.Contains(msg, "DTSTART;VALUE=DATE:20260314") {
		t.Fatalf("expected all-day DTSTART")
	}
}
