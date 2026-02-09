package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewKeyMapAppliesOverridesToLegend(t *testing.T) {
	keyMap := NewKeyMap(map[string][]string{
		"up":   []string{"k"},
		"run":  []string{"space"},
		"quit": []string{"q"},
	})

	help := keyMap.ShortHelp(false)
	if !strings.Contains(help, "k: up") {
		t.Fatalf("expected legend to include overridden up key, got %q", help)
	}
	if !strings.Contains(help, "space: run tests") {
		t.Fatalf("expected legend to include overridden run key, got %q", help)
	}
	if !strings.Contains(help, "q: quit") {
		t.Fatalf("expected legend to include overridden quit key, got %q", help)
	}
}

func TestNewKeyMapCompactLegend(t *testing.T) {
	keyMap := DefaultKeyMap()
	help := keyMap.ShortHelp(true)

	if strings.Contains(help, "select all") {
		t.Fatalf("compact help should omit select all actions: %q", help)
	}
	if !strings.Contains(help, "run tests") {
		t.Fatalf("compact help should include run action: %q", help)
	}
}

func TestNewKeyMapNormalizesOverrideKeys(t *testing.T) {
	keyMap := NewKeyMap(map[string][]string{
		"run":  []string{"Return"},
		"quit": []string{"escape", "ctrl-c"},
	})

	if keyLabel := keyMap.Run.Help().Key; keyLabel != "enter" {
		t.Fatalf("expected run key to normalize to enter, got %q", keyLabel)
	}
	if keyLabel := keyMap.Quit.Help().Key; keyLabel != "esc/^c" {
		t.Fatalf("expected quit keys to normalize to esc/^c, got %q", keyLabel)
	}
}

func TestNewKeyMapIgnoresUnknownActions(t *testing.T) {
	keyMap := NewKeyMap(map[string][]string{
		"unknown": []string{"x"},
	})

	if keyLabel := keyMap.Up.Help().Key; keyLabel != "↑/^k" {
		t.Fatalf("expected default up help label, got %q", keyLabel)
	}
}

func TestNewKeyMapSpaceAliasMatchesKeypress(t *testing.T) {
	keyMap := NewKeyMap(map[string][]string{
		"run": []string{"space"},
	})

	if got := keyMap.Run.Keys(); len(got) != 1 || got[0] != " " {
		t.Fatalf("expected run key to normalize to a literal space, got %v", got)
	}

	msg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	if !key.Matches(msg, keyMap.Run) {
		t.Fatalf("expected literal space keypress to match run binding")
	}

	if keyLabel := keyMap.Run.Help().Key; keyLabel != "space" {
		t.Fatalf("expected help label to display as space, got %q", keyLabel)
	}
}
