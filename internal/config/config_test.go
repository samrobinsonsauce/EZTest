package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func prepareConfigPath(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	configPath, err := GetAppConfigPath()
	if err != nil {
		t.Fatalf("failed to resolve config path: %v", err)
	}
	return configPath
}

func TestLoadAppSettingsDefaultsWhenMissing(t *testing.T) {
	_ = prepareConfigPath(t)

	settings, err := LoadAppSettings()
	if err != nil {
		t.Fatalf("LoadAppSettings returned error: %v", err)
	}

	if settings.Theme != "default" {
		t.Fatalf("expected default theme, got %q", settings.Theme)
	}
	if len(settings.Keybinds) != 0 {
		t.Fatalf("expected no keybind overrides, got %v", settings.Keybinds)
	}
	if !settings.UI.Animations {
		t.Fatalf("expected animations enabled by default")
	}
	if settings.UI.CompactHelp {
		t.Fatalf("expected compact help disabled by default")
	}
	if settings.Run.FailFast {
		t.Fatalf("expected fail_fast disabled by default")
	}
}

func TestGetAppConfigPathUsesEztestDir(t *testing.T) {
	configPath := prepareConfigPath(t)
	expected := filepath.Join(".config", "eztest", "config.json")
	if !strings.Contains(configPath, expected) {
		t.Fatalf("expected config path to include %q, got %q", expected, configPath)
	}
}

func TestLoadAppSettingsReadsAndNormalizes(t *testing.T) {
	configPath := prepareConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configJSON := `{
  "theme": "  Catppucin  ",
  "keybinds": {
    "UP": [" K ", "up", "CTRL-K", "", "k"],
    "run": ["Return"],
    "quit": ["escape", "ctrl-c"]
  },
  "ui": {
    "animations": false,
    "compact_help": true
  },
  "run": {
    "fail_fast": true
  }
}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	settings, err := LoadAppSettings()
	if err != nil {
		t.Fatalf("LoadAppSettings returned error: %v", err)
	}

	if settings.Theme != "catppucin" {
		t.Fatalf("expected normalized theme %q, got %q", "catppucin", settings.Theme)
	}

	if got, want := settings.Keybinds["up"], []string{"k", "up", "ctrl+k"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected up keybinds: got %v want %v", got, want)
	}
	if got, want := settings.Keybinds["run"], []string{"enter"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected run keybinds: got %v want %v", got, want)
	}
	if got, want := settings.Keybinds["quit"], []string{"esc", "ctrl+c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected quit keybinds: got %v want %v", got, want)
	}
	if settings.UI.Animations {
		t.Fatalf("expected animations to be disabled")
	}
	if !settings.UI.CompactHelp {
		t.Fatalf("expected compact help to be enabled")
	}
	if !settings.Run.FailFast {
		t.Fatalf("expected fail_fast to be enabled")
	}
}

func TestLoadAppSettingsInvalidJSONFallsBack(t *testing.T) {
	configPath := prepareConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"theme":`), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	settings, err := LoadAppSettings()
	if err == nil {
		t.Fatalf("expected parse error for invalid config")
	}
	if settings.Theme != "default" {
		t.Fatalf("expected fallback default theme, got %q", settings.Theme)
	}
	if !settings.UI.Animations || settings.UI.CompactHelp {
		t.Fatalf("expected default UI settings, got %+v", settings.UI)
	}
	if settings.Run.FailFast {
		t.Fatalf("expected default fail_fast=false, got true")
	}
}

func TestLoadAppSettingsFallsBackToLegacyPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	legacyPath := filepath.Join(home, ".config", "ezt", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatalf("failed to create legacy config dir: %v", err)
	}

	configJSON := `{"theme":"gruvbox","ui":{"animations":false}}`
	if err := os.WriteFile(legacyPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("failed to write legacy config file: %v", err)
	}

	settings, err := LoadAppSettings()
	if err != nil {
		t.Fatalf("LoadAppSettings returned error: %v", err)
	}

	if settings.Theme != "gruvbox" {
		t.Fatalf("expected to load theme from legacy config, got %q", settings.Theme)
	}
	if settings.UI.Animations {
		t.Fatalf("expected animations to be loaded from legacy config")
	}
	if settings.Run.FailFast {
		t.Fatalf("expected fail_fast default false when omitted")
	}
}

func TestLoadStateFallsBackToLegacyPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	projectDir := "/tmp/my_project"
	legacyPath := filepath.Join(home, ".config", "ezt", "state.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0755); err != nil {
		t.Fatalf("failed to create legacy state dir: %v", err)
	}

	stateJSON := `{"project_selections":{"` + projectDir + `":["test/foo_test.exs"]}}`
	if err := os.WriteFile(legacyPath, []byte(stateJSON), 0644); err != nil {
		t.Fatalf("failed to write legacy state file: %v", err)
	}

	selections, err := GetProjectSelections(projectDir)
	if err != nil {
		t.Fatalf("GetProjectSelections returned error: %v", err)
	}

	if got, want := selections, []string{"test/foo_test.exs"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected selections from legacy state: got %v want %v", got, want)
	}

	failures, err := GetProjectFailures(projectDir)
	if err != nil {
		t.Fatalf("GetProjectFailures returned error: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures in legacy state without project_failures field, got %v", failures)
	}
}

func TestSaveAndGetProjectFailures(t *testing.T) {
	_ = prepareConfigPath(t)

	projectDir := "/tmp/another_project"
	input := []string{"test/foo_test.exs", "test/bar_test.exs"}
	if err := SaveProjectFailures(projectDir, input); err != nil {
		t.Fatalf("SaveProjectFailures returned error: %v", err)
	}

	got, err := GetProjectFailures(projectDir)
	if err != nil {
		t.Fatalf("GetProjectFailures returned error: %v", err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("unexpected failures: got %v want %v", got, input)
	}

	empty, err := GetProjectFailures("/tmp/unknown")
	if err != nil {
		t.Fatalf("GetProjectFailures for unknown project returned error: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty failures for unknown project, got %v", empty)
	}
}
