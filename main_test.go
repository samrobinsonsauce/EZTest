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
