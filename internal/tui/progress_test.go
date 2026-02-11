package tui

import (
	"strings"
	"testing"
)

func TestFormatProgressLabel(t *testing.T) {
	if got := formatProgressLabel(6, 20); got != "6/20  30%" {
		t.Fatalf("unexpected progress label for known total: %q", got)
	}
	if got := formatProgressLabel(3, 0); got != "3 seen" {
		t.Fatalf("unexpected progress label for unknown total: %q", got)
	}
}

func TestRenderCanonicalProgressBarKnownTotal(t *testing.T) {
	ApplyTheme("default")

	raw := renderCanonicalProgressBar(12, 20, 0)
	bar := ansiEscapePattern.ReplaceAllString(raw, "")

	if got, want := len([]rune(bar)), canonicalProgressBarWidth; got != want {
		t.Fatalf("unexpected bar width: got %d want %d (%q)", got, want, bar)
	}
	if got, want := strings.Count(bar, "█"), 14; got != want {
		t.Fatalf("unexpected filled segments: got %d want %d (%q)", got, want, bar)
	}
	if got, want := strings.Count(bar, "░"), 10; got != want {
		t.Fatalf("unexpected empty segments: got %d want %d (%q)", got, want, bar)
	}
}

func TestRenderCanonicalProgressBarUnknownTotal(t *testing.T) {
	ApplyTheme("default")

	raw := renderCanonicalProgressBar(0, 0, 3)
	bar := ansiEscapePattern.ReplaceAllString(raw, "")

	if got, want := len([]rune(bar)), canonicalProgressBarWidth; got != want {
		t.Fatalf("unexpected bar width: got %d want %d (%q)", got, want, bar)
	}
	if got, want := strings.Count(bar, "▣"), 1; got != want {
		t.Fatalf("expected exactly one pulse marker: got %d (%q)", got, bar)
	}
}
