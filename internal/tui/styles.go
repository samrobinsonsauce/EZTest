package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#10B981")
	mutedColor     = lipgloss.Color("#6B7280")
	textColor      = lipgloss.Color("#F9FAFB")
	dimTextColor   = lipgloss.Color("#9CA3AF")
	borderColor    = lipgloss.Color("#374151")
	selectedBg     = lipgloss.Color("#1F2937")
)

var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(textColor)

	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(dimTextColor).
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(selectedBg).
				PaddingLeft(2)

	checkboxChecked = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			SetString("[✓]")

	checkboxUnchecked = lipgloss.NewStyle().
				Foreground(mutedColor).
				SetString("[ ]")

	cursorStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			SetString("▸")

	noCursorStyle = lipgloss.NewStyle().
			SetString(" ")

	statusStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	noResultsStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(1, 2)
)
