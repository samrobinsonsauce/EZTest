package tui

import (
	"reflect"
	"testing"
)

func TestExtractFailedFilesParsesFailureOutput(t *testing.T) {
	output := `
1) test example fails (MyApp.UserTest)
   test/user_test.exs:42
   Assertion with == failed

Finished in 0.1 seconds
2 tests, 1 failure
`

	runFiles := []string{"test/user_test.exs", "test/other_test.exs"}
	got := extractFailedFiles(output, runFiles)
	want := []string{"test/user_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractFailedFiles() = %v, want %v", got, want)
	}
}

func TestExtractFailedFilesHandlesAbsolutePathsAndAnsi(t *testing.T) {
	output := "\x1b[31mFailure\x1b[0m /tmp/project/test/api_test.exs:12"
	runFiles := []string{"test/api_test.exs", "test/another_test.exs"}

	got := extractFailedFiles(output, runFiles)
	want := []string{"test/api_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractFailedFiles() = %v, want %v", got, want)
	}
}

func TestExtractFailedFilesReturnsEmptyWhenNoMatches(t *testing.T) {
	output := "All tests passed."
	runFiles := []string{"test/api_test.exs"}
	got := extractFailedFiles(output, runFiles)
	if len(got) != 0 {
		t.Fatalf("expected no failed files, got %v", got)
	}
}
