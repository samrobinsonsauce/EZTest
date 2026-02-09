package tui

import (
	"strings"
	"testing"

	"github.com/samrobinsonsauce/eztest/internal/testfile"
)

func TestRenderItemShowsFailedMarker(t *testing.T) {
	ApplyTheme("default")

	item := Item{
		TestFile: testfile.TestFile{Path: "test/failing_test.exs"},
		Selected: true,
		Failed:   true,
	}

	rendered := RenderItem(item, 0, 0, 80, 0, false)
	if !strings.Contains(rendered, "✗") {
		t.Fatalf("expected failed marker in rendered item, got %q", rendered)
	}
}

func TestRenderItemHidesFailedMarkerForPassingFile(t *testing.T) {
	ApplyTheme("default")

	item := Item{
		TestFile: testfile.TestFile{Path: "test/passing_test.exs"},
		Selected: true,
		Failed:   false,
	}

	rendered := RenderItem(item, 0, 0, 80, 0, false)
	if strings.Contains(rendered, "✗") {
		t.Fatalf("did not expect failed marker in rendered item, got %q", rendered)
	}
}
