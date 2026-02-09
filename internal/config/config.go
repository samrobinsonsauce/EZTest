package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDirName       = "eztest"
	legacyConfigDirName = "ezt"
	stateFileName       = "state.json"
	appFileName         = "config.json"
)

type State struct {
	ProjectSelections map[string][]string `json:"project_selections"`
}

type AppSettings struct {
	Theme    string              `json:"theme"`
	Keybinds map[string][]string `json:"keybinds"`
	UI       UISettings          `json:"ui"`
}

type UISettings struct {
	Animations  bool `json:"animations"`
	CompactHelp bool `json:"compact_help"`
}

type rawAppSettings struct {
	Theme    string              `json:"theme"`
	Keybinds map[string][]string `json:"keybinds"`
	UI       rawUISettings       `json:"ui"`
}

type rawUISettings struct {
	Animations  *bool `json:"animations"`
	CompactHelp *bool `json:"compact_help"`
}

func getConfigDir() (string, error) {
	baseDir, err := getBaseConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, configDirName), nil
}

func getLegacyConfigDir() (string, error) {
	baseDir, err := getBaseConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, legacyConfigDirName), nil
}

func getBaseConfigDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

func getStatePath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

func GetAppConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appFileName), nil
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		Theme:    "default",
		Keybinds: map[string][]string{},
		UI: UISettings{
			Animations:  true,
			CompactHelp: false,
		},
	}
}

func LoadState() (*State, error) {
	configPath, err := getStatePath()
	if err != nil {
		return &State{ProjectSelections: make(map[string][]string)}, nil
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		legacyPath, legacyErr := getLegacyStatePath()
		if legacyErr != nil {
			return &State{ProjectSelections: make(map[string][]string)}, nil
		}
		data, err = os.ReadFile(legacyPath)
		if os.IsNotExist(err) {
			return &State{ProjectSelections: make(map[string][]string)}, nil
		}
	}
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return &State{ProjectSelections: make(map[string][]string)}, nil
	}

	if state.ProjectSelections == nil {
		state.ProjectSelections = make(map[string][]string)
	}

	return &state, nil
}

func SaveState(state *State) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, stateFileName)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func LoadAppSettings() (AppSettings, error) {
	settings := DefaultAppSettings()

	configPath, err := GetAppConfigPath()
	if err != nil {
		return settings, nil
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		legacyPath, legacyErr := getLegacyAppConfigPath()
		if legacyErr != nil {
			return settings, nil
		}
		data, err = os.ReadFile(legacyPath)
		if os.IsNotExist(err) {
			return settings, nil
		}
	}
	if err != nil {
		return settings, err
	}

	var raw rawAppSettings
	if err := json.Unmarshal(data, &raw); err != nil {
		return settings, fmt.Errorf("invalid app config at %s: %w", configPath, err)
	}

	if strings.TrimSpace(raw.Theme) != "" {
		settings.Theme = strings.ToLower(strings.TrimSpace(raw.Theme))
	}

	settings.Keybinds = sanitizeKeybinds(raw.Keybinds)

	if raw.UI.Animations != nil {
		settings.UI.Animations = *raw.UI.Animations
	}
	if raw.UI.CompactHelp != nil {
		settings.UI.CompactHelp = *raw.UI.CompactHelp
	}

	return settings, nil
}

func sanitizeKeybinds(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return map[string][]string{}
	}

	out := make(map[string][]string, len(in))
	for action, keys := range in {
		normalizedAction := strings.ToLower(strings.TrimSpace(action))
		if normalizedAction == "" {
			continue
		}

		cleaned := sanitizeKeys(keys)
		if len(cleaned) == 0 {
			continue
		}

		out[normalizedAction] = cleaned
	}

	return out
}

func sanitizeKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))

	for _, k := range keys {
		normalized := normalizeKey(k)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	return out
}

func normalizeKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}

	replacements := map[string]string{
		"ctrl-":  "ctrl+",
		"alt-":   "alt+",
		"cmd-":   "cmd+",
		"return": "enter",
		"escape": "esc",
	}
	for from, to := range replacements {
		if strings.HasPrefix(key, from) {
			return to + strings.TrimSpace(strings.TrimPrefix(key, from))
		}
	}

	return key
}

func getLegacyStatePath() (string, error) {
	dir, err := getLegacyConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

func getLegacyAppConfigPath() (string, error) {
	dir, err := getLegacyConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appFileName), nil
}

func GetProjectSelections(projectDir string) ([]string, error) {
	state, err := LoadState()
	if err != nil {
		return nil, err
	}

	selections, ok := state.ProjectSelections[projectDir]
	if !ok {
		return []string{}, nil
	}

	return selections, nil
}

func SaveProjectSelections(projectDir string, selections []string) error {
	state, err := LoadState()
	if err != nil {
		state = &State{ProjectSelections: make(map[string][]string)}
	}

	state.ProjectSelections[projectDir] = selections

	return SaveState(state)
}
