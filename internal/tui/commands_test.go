package tui

import (
	"os"
	"path/filepath"
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

func TestExtractFailedFilesWithoutRunFilesCollectsObservedPaths(t *testing.T) {
	output := `
1) test fails (App.SomeTest)
   test/some_test.exs:42
   Assertion with == failed
`

	got := extractFailedFiles(output, nil)
	want := []string{"test/some_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractFailedFiles() with no run files = %v, want %v", got, want)
	}
}

func TestParseTestRunStatsIncludesDescribeCount(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "user_test.exs")

	content := `
describe "user" do
  test "a" do
  end
end

describe "auth" do
  test "b" do
  end
end
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	output := `
Finished in 0.08 seconds (0.00s async, 0.08s sync)
12 tests, 2 failures
`
	stats := parseTestRunStats(output, []string{testFile})

	if stats.Tests != 12 {
		t.Fatalf("expected 12 tests, got %d", stats.Tests)
	}
	if stats.Failures != 2 {
		t.Fatalf("expected 2 failures, got %d", stats.Failures)
	}
	if stats.DescribeCount != 2 {
		t.Fatalf("expected 2 describe blocks, got %d", stats.DescribeCount)
	}
	if stats.Duration != "0.08 seconds (0.00s async, 0.08s sync)" {
		t.Fatalf("unexpected duration: %q", stats.Duration)
	}
}

func TestParseFailureDetails(t *testing.T) {
	output := `
1) test creates user (MyApp.UserTest)
   test/user_test.exs:12
   Assertion with == failed
   code:  assert user.id == nil

2) test validates email (MyApp.UserTest)
   test/user_test.exs:30
   Expected truthy, got false

Finished in 0.03 seconds
2 tests, 2 failures
`

	failures := parseFailureDetails(output)
	if got, want := len(failures), 2; got != want {
		t.Fatalf("expected %d failure details, got %d", want, got)
	}

	first := failures[0]
	if first.Index != 1 || first.Name != "test creates user" || first.Module != "MyApp.UserTest" {
		t.Fatalf("unexpected first failure header: %+v", first)
	}
	if first.File != "test/user_test.exs" || first.Line != 12 {
		t.Fatalf("unexpected first failure location: %+v", first)
	}
	if first.Details == "" {
		t.Fatalf("expected first failure details to be captured")
	}
}

func TestProgressTrackerCountsOnlyProgressLines(t *testing.T) {
	tracker := newProgressTracker(false, 10)

	tracker.Consume([]byte("....F"))
	tracker.Consume([]byte("\n1) test something failed (MyApp.Test)\n"))
	tracker.Consume([]byte("   code: assert map.key == 1.0\n"))
	tracker.Consume([]byte(".."))

	pass, fail, other := tracker.Snapshot()
	if pass != 6 {
		t.Fatalf("expected 6 passing ticks, got %d", pass)
	}
	if fail != 1 {
		t.Fatalf("expected 1 failing tick, got %d", fail)
	}
	if other != 0 {
		t.Fatalf("expected 0 other ticks, got %d", other)
	}
}

func TestProgressTrackerCountsOtherMarkers(t *testing.T) {
	tracker := newProgressTracker(false, 10)
	tracker.Consume([]byte(".F*S?\n"))

	pass, fail, other := tracker.Snapshot()
	if pass != 1 || fail != 1 || other != 3 {
		t.Fatalf("unexpected progress counts: pass=%d fail=%d other=%d", pass, fail, other)
	}
}

func TestProgressTrackerCountsLeadingFailureImmediately(t *testing.T) {
	tracker := newProgressTracker(false, 10)
	tracker.Consume([]byte("F"))

	pass, fail, other := tracker.Snapshot()
	if pass != 0 || fail != 1 || other != 0 {
		t.Fatalf("expected immediate fail tick, got pass=%d fail=%d other=%d", pass, fail, other)
	}
}

func TestProgressTrackerIgnoresNonProgressLinesStartingWithF(t *testing.T) {
	tracker := newProgressTracker(false, 10)
	tracker.Consume([]byte("Finished in 0.1 seconds\n"))

	pass, fail, other := tracker.Snapshot()
	if pass != 0 || fail != 0 || other != 0 {
		t.Fatalf("expected no progress counts from non-progress lines, got pass=%d fail=%d other=%d", pass, fail, other)
	}
}

func TestProgressTrackerSyncWithSummaryUsesParsedFailureCount(t *testing.T) {
	tracker := newProgressTracker(false, 10)
	tracker.Consume([]byte("...FFF\n"))
	tracker.SyncWithSummary(TestRunStats{Failures: 1})

	_, fail, _ := tracker.Snapshot()
	if fail != 1 {
		t.Fatalf("expected fail counter synced to summary failure count, got %d", fail)
	}
}

func TestCountSelectedTestsScansMacros(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "sample_test.exs")

	content := `
defmodule SampleTest do
  use ExUnit.Case

  # test "commented out"
  test "works" do
  end

	property "generated" do
  end

  test("paren style", %{ctx: _ctx}) do
  end

  contest = :not_a_test
end
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if got := countSelectedTests([]string{file}); got != 3 {
		t.Fatalf("expected 3 tests from scan, got %d", got)
	}
}

func TestCountSelectedTestsDedupesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "duplicate_test.exs")

	content := `
defmodule DuplicateTest do
  use ExUnit.Case
  test "a" do
  end
end
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if got := countSelectedTests([]string{file, file}); got != 1 {
		t.Fatalf("expected duplicate files to be counted once, got %d", got)
	}
}

func TestProgressLabelWithKnownTotal(t *testing.T) {
	tracker := newProgressTracker(false, 20)
	tracker.Consume([]byte(".....F"))

	label := tracker.progressLabel(tracker.totalSeen())
	if label != "6/20  30%" {
		t.Fatalf("unexpected progress label: %q", label)
	}
}

func TestBuildMixTestArgsRespectsFailFast(t *testing.T) {
	files := []string{"test/a_test.exs", "test/b_test.exs"}

	got := buildMixTestArgs(files, RunOptions{FailFast: true}, false)
	want := []string{"test", "--max-failures", "1", "test/a_test.exs", "test/b_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected args with fail_fast: got %v want %v", got, want)
	}

	got = buildMixTestArgs(files, RunOptions{FailFast: false}, false)
	want = []string{"test", "--max-failures", "2147483647", "test/a_test.exs", "test/b_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected args without fail_fast: got %v want %v", got, want)
	}

	got = buildMixTestArgs(nil, RunOptions{FailFast: false}, true)
	want = []string{"test", "--max-failures", "2147483647", "--failed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected args for --failed: got %v want %v", got, want)
	}
}
