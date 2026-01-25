package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/samrobinsonsauce/eztest/internal/testfile"
)

type Item struct {
	TestFile testfile.TestFile
	Selected bool
}

func (i Item) FilterValue() string {
	return i.TestFile.Path
}

func RenderItem(item Item, index int, cursor int, width int, frame int) string {
	isCursor := index == cursor

	var checkbox string
	if item.Selected {
		checkmarks := []string{"[✓]", "[✓]", "[✓]", "[★]"}
		checkbox = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Render(checkmarks[frame%len(checkmarks)])
	} else {
		checkbox = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render("[ ]")
	}

	var cursorIndicator string
	if isCursor {
		cursors := []string{"▸", "▹", "▸", "▹"}
		cursorIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Render(cursors[frame%len(cursors)])
	} else {
		cursorIndicator = " "
	}

	path := item.TestFile.Path

	maxPathWidth := width - 8
	if maxPathWidth > 0 && len(path) > maxPathWidth {
		path = "..." + path[len(path)-maxPathWidth+3:]
	}

	line := cursorIndicator + " " + checkbox + " " + path

	var style lipgloss.Style
	if isCursor {
		style = selectedItemStyle.Width(width)
	} else {
		style = itemStyle.Width(width)
	}

	return style.Render(line)
}
