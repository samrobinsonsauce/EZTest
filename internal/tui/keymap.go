package tui

import "github.com/charmbracelet/bubbles/key"

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
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "ctrl+k"),
			key.WithHelp("↑/^k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "ctrl+j"),
			key.WithHelp("↓/^j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("^a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("^d", "deselect all"),
		),
		Run: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run tests"),
		),
		SaveQuit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("^s", "save & quit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() string {
	return "↑/^k ↓/^j: navigate • tab: select • ^a/^d: all/none • enter: run • ^s: save • esc: quit"
}
