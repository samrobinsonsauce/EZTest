package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

const (
	actionUp          = "up"
	actionDown        = "down"
	actionSelect      = "select"
	actionSelectAll   = "select_all"
	actionDeselectAll = "deselect_all"
	actionRun         = "run"
	actionSaveQuit    = "save_quit"
	actionQuit        = "quit"
)

type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
	Run         key.Binding
	SaveQuit    key.Binding
	Quit        key.Binding
}

func DefaultKeyMap() KeyMap {
	return NewKeyMap(nil)
}

func NewKeyMap(overrides map[string][]string) KeyMap {
	bindings := defaultBindings()
	for action, keys := range overrides {
		defaultKeys, ok := bindings[action]
		if !ok {
			continue
		}

		cleaned := cleanBindingKeys(keys)
		if len(cleaned) == 0 {
			bindings[action] = defaultKeys
			continue
		}

		bindings[action] = cleaned
	}

	return KeyMap{
		Up:          makeBinding(bindings[actionUp], "up"),
		Down:        makeBinding(bindings[actionDown], "down"),
		Select:      makeBinding(bindings[actionSelect], "select"),
		SelectAll:   makeBinding(bindings[actionSelectAll], "select all"),
		DeselectAll: makeBinding(bindings[actionDeselectAll], "deselect all"),
		Run:         makeBinding(bindings[actionRun], "run tests"),
		SaveQuit:    makeBinding(bindings[actionSaveQuit], "save & quit"),
		Quit:        makeBinding(bindings[actionQuit], "quit"),
	}
}

func (k KeyMap) ShortHelp(compact bool) string {
	entries := []key.Binding{
		k.Up,
		k.Down,
		k.Select,
		k.SelectAll,
		k.DeselectAll,
		k.Run,
		k.SaveQuit,
		k.Quit,
	}
	if compact {
		entries = []key.Binding{k.Up, k.Down, k.Select, k.Run, k.Quit}
	}

	parts := make([]string, 0, len(entries))
	for _, binding := range entries {
		help := binding.Help()
		if help.Key == "" || help.Desc == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", help.Key, help.Desc))
	}

	return strings.Join(parts, " • ")
}

func defaultBindings() map[string][]string {
	return map[string][]string{
		actionUp:          []string{"up", "ctrl+k"},
		actionDown:        []string{"down", "ctrl+j"},
		actionSelect:      []string{"tab"},
		actionSelectAll:   []string{"ctrl+a"},
		actionDeselectAll: []string{"ctrl+d"},
		actionRun:         []string{"enter"},
		actionSaveQuit:    []string{"ctrl+s"},
		actionQuit:        []string{"ctrl+c", "esc"},
	}
}

func cleanBindingKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	cleaned := make([]string, 0, len(keys))
	for _, k := range keys {
		normalized := normalizeBindingKey(k)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		cleaned = append(cleaned, normalized)
	}
	return cleaned
}

func normalizeBindingKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	switch key {
	case "":
		return ""
	case "return":
		return "enter"
	case "escape":
		return "esc"
	case "space", "spacebar":
		return " "
	}

	prefixes := map[string]string{
		"ctrl-": "ctrl+",
		"alt-":  "alt+",
		"cmd-":  "cmd+",
	}
	for from, to := range prefixes {
		if strings.HasPrefix(key, from) {
			return to + strings.TrimSpace(strings.TrimPrefix(key, from))
		}
	}

	return key
}

func makeBinding(keys []string, helpDescription string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(formatHelpKeys(keys), helpDescription),
	)
}

func formatHelpKeys(keys []string) string {
	if len(keys) == 0 {
		return ""
	}

	labels := make([]string, 0, len(keys))
	for _, k := range keys {
		labels = append(labels, formatHelpKey(k))
	}
	return strings.Join(labels, "/")
}

func formatHelpKey(keyName string) string {
	switch keyName {
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	case " ", "space":
		return "space"
	}

	if strings.HasPrefix(keyName, "ctrl+") {
		return "^" + strings.TrimPrefix(keyName, "ctrl+")
	}

	return keyName
}
