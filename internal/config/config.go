package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = "state.json"

type State struct {
	ProjectSelections map[string][]string `json:"project_selections"`
}

func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "ezt"), nil
}

func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func LoadState() (*State, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &State{ProjectSelections: make(map[string][]string)}, nil
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &State{ProjectSelections: make(map[string][]string)}, nil
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

	configPath := filepath.Join(configDir, configFileName)

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
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
