package tui

import (
	"github.com/samrobinsonsauce/eztest/internal/testfile"
)

type Item struct {
	TestFile testfile.TestFile
	Selected bool
}

func (i Item) FilterValue() string {
	return i.TestFile.Path
}

func RenderItem(item Item, index int, cursor int, width int, frame int, animate bool) string {
	isCursor := index == cursor

	var checkbox string
	if item.Selected {
		checkmark := "[✓]"
		if animate {
			checkmarks := []string{"[✓]", "[✓]", "[✓]", "[★]"}
			checkmark = checkmarks[frame%len(checkmarks)]
		}
		checkbox = checkboxCheckedStyle.Render(checkmark)
	} else {
		checkbox = checkboxUncheckedStyle.Render("[ ]")
	}

	var cursorIndicator string
	if isCursor {
		cursorChar := "▸"
		if animate {
			cursors := []string{"▸", "▹", "▸", "▹"}
			cursorChar = cursors[frame%len(cursors)]
		}
		cursorIndicator = cursorStyle.Render(cursorChar)
	} else {
		cursorIndicator = noCursorStyle.Render(" ")
	}

	path := item.TestFile.Path

	maxPathWidth := width - 8
	if maxPathWidth > 0 && len(path) > maxPathWidth {
		path = "..." + path[len(path)-maxPathWidth+3:]
	}

	line := cursorIndicator + " " + checkbox + " " + path

	if isCursor {
		return selectedItemStyle.Width(width).Render(line)
	} else {
		return itemStyle.Width(width).Render(line)
	}
}
