package tui

import (
	"testing"

	"github.com/samrobinsonsauce/eztest/internal/config"
	"github.com/samrobinsonsauce/eztest/internal/testfile"
)

func testModelForFailures() Model {
	files := []testfile.TestFile{
		{Path: "test/user_test.exs"},
		{Path: "test/api_test.exs"},
		{Path: "test/auth_test.exs"},
	}

	return NewModel(
		files,
		"/tmp/project",
		[]string{"test/user_test.exs"},
		[]string{"test/api_test.exs", "test/auth_test.exs"},
		DefaultKeyMap(),
		config.UISettings{Animations: false},
	)
}

func TestNewModelMarksFailedItems(t *testing.T) {
	m := testModelForFailures()

	failed := map[string]bool{}
	for _, item := range m.allItems {
		failed[item.TestFile.Path] = item.Failed
	}

	if failed["test/user_test.exs"] {
		t.Fatalf("expected non-failed file to remain unmarked")
	}
	if !failed["test/api_test.exs"] {
		t.Fatalf("expected api_test to be marked as failed")
	}
	if !failed["test/auth_test.exs"] {
		t.Fatalf("expected auth_test to be marked as failed")
	}
}

func TestUpdateFilterFailedTokenShowsOnlyFailed(t *testing.T) {
	m := testModelForFailures()
	m.searchInput.SetValue("@failed")
	m.updateFilter()

	if got, want := len(m.filteredItems), 2; got != want {
		t.Fatalf("expected %d filtered items, got %d", want, got)
	}

	for _, item := range m.filteredItems {
		if !item.Failed {
			t.Fatalf("expected only failed items with @failed filter, found %q", item.TestFile.Path)
		}
	}
}

func TestUpdateFilterFailedTokenCombinedWithQuery(t *testing.T) {
	m := testModelForFailures()
	m.searchInput.SetValue("@failed api")
	m.updateFilter()

	if got, want := len(m.filteredItems), 1; got != want {
		t.Fatalf("expected %d filtered items, got %d", want, got)
	}
	if got, want := m.filteredItems[0].TestFile.Path, "test/api_test.exs"; got != want {
		t.Fatalf("unexpected filtered file: got %q want %q", got, want)
	}
}
