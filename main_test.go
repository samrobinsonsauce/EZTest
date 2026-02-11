package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/samrobinsonsauce/eztest/internal/config"
	"github.com/samrobinsonsauce/eztest/internal/tui"
)

func setupConfigEnv(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
}

func TestRunAndPersistFailuresSuccess(t *testing.T) {
	setupConfigEnv(t)

	original := executeMixTest
	executeMixTest = func(files []string) (tui.TestRunOutcome, error) {
		return tui.TestRunOutcome{FailedFiles: []string{"test/a_test.exs"}}, nil
	}
	t.Cleanup(func() {
		executeMixTest = original
	})

	code := runAndPersistFailures("/tmp/project", []string{"test/a_test.exs", "test/b_test.exs"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	got, err := config.GetProjectFailures("/tmp/project")
	if err != nil {
		t.Fatalf("GetProjectFailures returned error: %v", err)
	}
	want := []string{"test/a_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected persisted failures: got %v want %v", got, want)
	}
}

func TestRunAndPersistFailuresSkipsPersistenceOnGenericError(t *testing.T) {
	setupConfigEnv(t)
	project := "/tmp/project2"

	if err := config.SaveProjectFailures(project, []string{"test/existing_test.exs"}); err != nil {
		t.Fatalf("SaveProjectFailures returned error: %v", err)
	}

	original := executeMixTest
	executeMixTest = func(files []string) (tui.TestRunOutcome, error) {
		return tui.TestRunOutcome{FailedFiles: []string{"test/new_failure_test.exs"}}, errors.New("boom")
	}
	t.Cleanup(func() {
		executeMixTest = original
	})

	code := runAndPersistFailures(project, []string{"test/new_failure_test.exs"})
	if code != 1 {
		t.Fatalf("expected exit code 1 for generic error, got %d", code)
	}

	got, err := config.GetProjectFailures(project)
	if err != nil {
		t.Fatalf("GetProjectFailures returned error: %v", err)
	}
	want := []string{"test/existing_test.exs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected existing failures to remain unchanged, got %v want %v", got, want)
	}
}

func TestFailuresForProjectReadsSavedFailures(t *testing.T) {
	setupConfigEnv(t)
	project := "/tmp/project3"
	want := []string{"test/failure_test.exs"}

	if err := config.SaveProjectFailures(project, want); err != nil {
		t.Fatalf("SaveProjectFailures returned error: %v", err)
	}

	if got := failuresForProject(project); !reflect.DeepEqual(got, want) {
		t.Fatalf("failuresForProject() = %v, want %v", got, want)
	}
}

func TestRunWithResultsScreenRerunFailed(t *testing.T) {
	setupConfigEnv(t)

	originalExec := executeMixTest
	originalExecFailed := executeMixTestFailed
	originalPrompt := promptRerunAction
	originalPrint := printResultsScreen
	originalInteractive := isInteractiveSession
	t.Cleanup(func() {
		executeMixTest = originalExec
		executeMixTestFailed = originalExecFailed
		promptRerunAction = originalPrompt
		printResultsScreen = originalPrint
		isInteractiveSession = originalInteractive
	})

	selectedRunCount := 0
	executedWith := make([][]string, 0, 1)
	executeMixTest = func(files []string) (tui.TestRunOutcome, error) {
		snapshot := append([]string(nil), files...)
		executedWith = append(executedWith, snapshot)
		selectedRunCount++
		return tui.TestRunOutcome{
			FailedFiles: []string{"test/failed_test.exs"},
			Stats:       tui.TestRunStats{Failures: 1},
		}, nil
	}
	failedRunCount := 0
	executeMixTestFailed = func() (tui.TestRunOutcome, error) {
		failedRunCount++
		return tui.TestRunOutcome{
			FailedFiles: []string{},
			Stats:       tui.TestRunStats{Failures: 0},
		}, nil
	}

	promptCalls := 0
	promptRerunAction = func(outcome tui.TestRunOutcome) tui.RerunAction {
		promptCalls++
		if promptCalls == 1 {
			return tui.RerunActionFailed
		}
		return tui.RerunActionQuit
	}
	printResultsScreen = func(outcome tui.TestRunOutcome, exitCode int) {}
	isInteractiveSession = func() bool { return true }

	exitCode := runWithResultsScreen("/tmp/project-rerun-failed", []string{"test/first_test.exs", "test/second_test.exs"}, false)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if selectedRunCount != 1 {
		t.Fatalf("expected selected run to execute once, got %d", selectedRunCount)
	}
	if failedRunCount != 1 {
		t.Fatalf("expected failed rerun to execute once, got %d", failedRunCount)
	}
	if got, want := executedWith[0], []string{"test/first_test.exs", "test/second_test.exs"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected first execution files: got %v want %v", got, want)
	}
}

func TestRunWithResultsScreenRerunAll(t *testing.T) {
	setupConfigEnv(t)

	originalExec := executeMixTest
	originalExecFailed := executeMixTestFailed
	originalPrompt := promptRerunAction
	originalPrint := printResultsScreen
	originalInteractive := isInteractiveSession
	t.Cleanup(func() {
		executeMixTest = originalExec
		executeMixTestFailed = originalExecFailed
		promptRerunAction = originalPrompt
		printResultsScreen = originalPrint
		isInteractiveSession = originalInteractive
	})

	callCount := 0
	executedWith := make([][]string, 0, 2)
	executeMixTest = func(files []string) (tui.TestRunOutcome, error) {
		snapshot := append([]string(nil), files...)
		executedWith = append(executedWith, snapshot)
		callCount++
		return tui.TestRunOutcome{}, nil
	}
	executeMixTestFailed = func() (tui.TestRunOutcome, error) {
		t.Fatalf("did not expect executeMixTestFailed to be called")
		return tui.TestRunOutcome{}, nil
	}

	promptCalls := 0
	promptRerunAction = func(outcome tui.TestRunOutcome) tui.RerunAction {
		promptCalls++
		if promptCalls == 1 {
			return tui.RerunActionAll
		}
		return tui.RerunActionQuit
	}
	printResultsScreen = func(outcome tui.TestRunOutcome, exitCode int) {}
	isInteractiveSession = func() bool { return true }

	initial := []string{"test/all_test.exs"}
	exitCode := runWithResultsScreen("/tmp/project-rerun-all", initial, false)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 executions, got %d", callCount)
	}
	if !reflect.DeepEqual(executedWith[0], initial) || !reflect.DeepEqual(executedWith[1], initial) {
		t.Fatalf("expected rerun-all to rerun same file set, got %v", executedWith)
	}
}
