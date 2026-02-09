package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type palette struct {
	Primary    string
	Secondary  string
	Muted      string
	Text       string
	DimText    string
	Border     string
	SelectedBg string
	Error      string
}

type Theme struct {
	Name    string
	Palette palette
}

var themePresets = map[string]palette{
	"default": {
		Primary:    "#7C3AED",
		Secondary:  "#10B981",
		Muted:      "#6B7280",
		Text:       "#F9FAFB",
		DimText:    "#9CA3AF",
		Border:     "#374151",
		SelectedBg: "#1F2937",
		Error:      "#EF4444",
	},
	"gruvbox": {
		Primary:    "#FABD2F",
		Secondary:  "#B8BB26",
		Muted:      "#928374",
		Text:       "#EBDBB2",
		DimText:    "#A89984",
		Border:     "#504945",
		SelectedBg: "#3C3836",
		Error:      "#FB4934",
	},
	"catppuccin": {
		Primary:    "#89B4FA",
		Secondary:  "#A6E3A1",
		Muted:      "#7F849C",
		Text:       "#CDD6F4",
		DimText:    "#A6ADC8",
		Border:     "#45475A",
		SelectedBg: "#313244",
		Error:      "#F38BA8",
	},
}

var themeAliases = map[string]string{
	"":           "default",
	"default":    "default",
	"gruvbox":    "gruvbox",
	"gruv":       "gruvbox",
	"catppuccin": "catppuccin",
	"catppucin":  "catppuccin",
	"catpuccin":  "catppuccin",
}

var (
	currentThemeName = "default"
	primaryColor     lipgloss.Color
	secondaryColor   lipgloss.Color
	mutedColor       lipgloss.Color
	textColor        lipgloss.Color
	dimTextColor     lipgloss.Color
	borderColor      lipgloss.Color
	selectedBg       lipgloss.Color
)

var (
	appStyle lipgloss.Style

	titleStyle lipgloss.Style

	searchPromptStyle lipgloss.Style

	searchInputStyle lipgloss.Style

	searchBoxStyle lipgloss.Style

	listStyle lipgloss.Style

	itemStyle lipgloss.Style

	selectedItemStyle lipgloss.Style

	checkboxCheckedStyle lipgloss.Style

	checkboxUncheckedStyle lipgloss.Style

	failedMarkerStyle lipgloss.Style

	cursorStyle lipgloss.Style

	noCursorStyle lipgloss.Style

	statusStyle lipgloss.Style

	helpStyle lipgloss.Style

	errorStyle lipgloss.Style

	noResultsStyle lipgloss.Style

	bannerStyle    lipgloss.Style
	logoStyle      lipgloss.Style
	fileCountStyle lipgloss.Style
	fileListStyle  lipgloss.Style
	dividerStyle   lipgloss.Style
)

func init() {
	ApplyTheme("default")
}

func ApplyTheme(name string) string {
	theme := ResolveTheme(name)
	currentThemeName = theme.Name
	applyPalette(theme.Palette)
	return currentThemeName
}

func CurrentThemeName() string {
	return currentThemeName
}

func ResolveTheme(name string) Theme {
	normalized := normalizeThemeName(name)
	canonical, ok := themeAliases[normalized]
	if !ok {
		canonical = "default"
	}

	p, ok := themePresets[canonical]
	if !ok {
		canonical = "default"
		p = themePresets[canonical]
	}

	return Theme{Name: canonical, Palette: p}
}

func normalizeThemeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	return replacer.Replace(name)
}

func applyPalette(p palette) {
	primaryColor = lipgloss.Color(p.Primary)
	secondaryColor = lipgloss.Color(p.Secondary)
	mutedColor = lipgloss.Color(p.Muted)
	textColor = lipgloss.Color(p.Text)
	dimTextColor = lipgloss.Color(p.DimText)
	borderColor = lipgloss.Color(p.Border)
	selectedBg = lipgloss.Color(p.SelectedBg)
	errorColor := lipgloss.Color(p.Error)

	appStyle = lipgloss.NewStyle().
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

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

	checkboxCheckedStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true)

	checkboxUncheckedStyle = lipgloss.NewStyle().
		Foreground(mutedColor)

	failedMarkerStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	cursorStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	noCursorStyle = lipgloss.NewStyle()

	statusStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		MarginTop(1)

	helpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)

	errorStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	noResultsStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Padding(1, 2)

	bannerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 2).
		MarginBottom(1)

	logoStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)

	fileCountStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(secondaryColor)

	fileListStyle = lipgloss.NewStyle().
		Foreground(dimTextColor).
		MarginLeft(2)

	dividerStyle = lipgloss.NewStyle().
		Foreground(borderColor)
}
