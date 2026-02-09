package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samrobinsonsauce/eztest/internal/config"
	"github.com/samrobinsonsauce/eztest/internal/testfile"
)

type Model struct {
	allItems      []Item
	filteredItems []Item
	projectDir    string
	cursor        int
	searchInput   textinput.Model
	keyMap        KeyMap
	animations    bool
	compactHelp   bool
	width         int
	height        int
	frame         int
	filesToRun    []string
	quitting      bool
}

type tickMsg time.Time

func NewModel(testFiles []testfile.TestFile, projectDir string, selections []string, failures []string, keyMap KeyMap, ui config.UISettings) Model {
	selectedSet := make(map[string]bool)
	for _, s := range selections {
		selectedSet[s] = true
	}
	failedSet := make(map[string]bool)
	for _, f := range failures {
		failedSet[f] = true
	}

	items := make([]Item, len(testFiles))
	for i, tf := range testFiles {
		items[i] = Item{
			TestFile: tf,
			Selected: selectedSet[tf.Path],
			Failed:   failedSet[tf.Path],
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Type to filter tests..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.PromptStyle = searchPromptStyle
	ti.TextStyle = searchInputStyle
	ti.Prompt = "🔍 "

	if len(keyMap.Up.Keys()) == 0 {
		keyMap = DefaultKeyMap()
	}

	return Model{
		allItems:      items,
		filteredItems: items,
		projectDir:    projectDir,
		cursor:        0,
		searchInput:   ti,
		keyMap:        keyMap,
		animations:    ui.Animations,
		compactHelp:   ui.CompactHelp,
		width:         80,
		height:        24,
		frame:         0,
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	if m.animations {
		cmds = append(cmds, tick())
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.animations {
			m.frame++
			return m, tick()
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.SaveQuit):
			selections := m.getSelectedFiles()
			_ = config.SaveProjectSelections(m.projectDir, selections)
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Down):
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Select):
			if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
				filteredItem := &m.filteredItems[m.cursor]
				filteredItem.Selected = !filteredItem.Selected

				for i := range m.allItems {
					if m.allItems[i].TestFile.Path == filteredItem.TestFile.Path {
						m.allItems[i].Selected = filteredItem.Selected
						break
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keyMap.SelectAll):
			for i := range m.filteredItems {
				m.filteredItems[i].Selected = true
				for j := range m.allItems {
					if m.allItems[j].TestFile.Path == m.filteredItems[i].TestFile.Path {
						m.allItems[j].Selected = true
						break
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keyMap.DeselectAll):
			for i := range m.allItems {
				m.allItems[i].Selected = false
			}
			for i := range m.filteredItems {
				m.filteredItems[i].Selected = false
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Run):
			m.filesToRun = m.getSelectedFiles()
			_ = config.SaveProjectSelections(m.projectDir, m.filesToRun)
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = msg.Width - 10
		return m, nil
	}

	prevValue := m.searchInput.Value()
	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	if m.searchInput.Value() != prevValue {
		m.updateFilter()
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateFilter() {
	query := strings.TrimSpace(m.searchInput.Value())
	failedOnly := false
	if strings.Contains(strings.ToLower(query), "@failed") {
		failedOnly = true
		query = strings.ReplaceAll(strings.ToLower(query), "@failed", "")
	}

	if strings.TrimSpace(query) == "" {
		m.filteredItems = make([]Item, 0, len(m.allItems))
		for _, item := range m.allItems {
			if failedOnly && !item.Failed {
				continue
			}
			m.filteredItems = append(m.filteredItems, item)
		}
	} else {
		tokens := strings.Fields(strings.ToLower(query))

		m.filteredItems = make([]Item, 0)
		for _, item := range m.allItems {
			if failedOnly && !item.Failed {
				continue
			}

			pathLower := strings.ToLower(item.TestFile.Path)

			allMatch := true
			for _, token := range tokens {
				if !fuzzyContains(pathLower, token) {
					allMatch = false
					break
				}
			}

			if allMatch {
				m.filteredItems = append(m.filteredItems, item)
			}
		}

		if len(m.filteredItems) > 1 {
			sortByRelevance(m.filteredItems, tokens)
		}
	}

	if m.cursor >= len(m.filteredItems) {
		m.cursor = max(0, len(m.filteredItems)-1)
	}
}

func fuzzyContains(text, pattern string) bool {
	if strings.Contains(text, pattern) {
		return true
	}

	patternIdx := 0
	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
		}
	}
	return patternIdx == len(pattern)
}

func sortByRelevance(items []Item, tokens []string) {
	type scored struct {
		item  Item
		score int
	}

	scoredItems := make([]scored, len(items))
	for i, item := range items {
		pathLower := strings.ToLower(item.TestFile.Path)
		score := 0

		for _, token := range tokens {
			if strings.Contains(pathLower, token) {
				score += 100
				if strings.Contains(pathLower, "/"+token) || strings.Contains(pathLower, "_"+token) {
					score += 50
				}
			} else {
				score += 10
			}
		}

		score -= len(item.TestFile.Path) / 10

		scoredItems[i] = scored{item: item, score: score}
	}

	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].score > scoredItems[j].score
	})

	for i, s := range scoredItems {
		items[i] = s.item
	}
}

func (m *Model) getSelectedFiles() []string {
	var selected []string
	for _, item := range m.allItems {
		if item.Selected {
			selected = append(selected, item.TestFile.Path)
		}
	}
	return selected
}

func (m Model) GetFilesToRun() []string {
	return m.filesToRun
}

func (m Model) IsQuitting() bool {
	return m.quitting && len(m.filesToRun) == 0
}

func (m Model) getAnimatedTitle() string {
	titleText := "EZTest - Elixir Test Selector"
	if !m.animations {
		return titleStyle.Render(titleText)
	}

	cursors := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	cursor := cursors[m.frame%len(cursors)]
	return titleStyle.Render(cursor + " " + titleText)
}

func (m Model) View() string {
	if m.quitting {
		if len(m.filesToRun) > 0 {
			return fmt.Sprintf("\n  Running %d test file(s)...\n\n", len(m.filesToRun))
		}
		return ""
	}

	var b strings.Builder

	// Animated title
	b.WriteString(m.getAnimatedTitle())
	b.WriteString("\n\n")

	b.WriteString(searchBoxStyle.Render(m.searchInput.View()))
	b.WriteString("\n\n")

	listHeight := m.height - 12
	if listHeight < 5 {
		listHeight = 5
	}

	listWidth := m.width - 6
	if listWidth < 40 {
		listWidth = 40
	}

	if len(m.filteredItems) == 0 {
		dots := ""
		if m.animations {
			dots = strings.Repeat(".", (m.frame/3)%4)
		}
		noResultsText := fmt.Sprintf("No matching test files%s", dots)
		noResults := noResultsStyle.Render(noResultsText)
		b.WriteString(listStyle.Width(listWidth).Height(listHeight).Render(noResults))
	} else {
		start := 0
		end := len(m.filteredItems)

		if len(m.filteredItems) > listHeight {
			halfHeight := listHeight / 2
			start = m.cursor - halfHeight
			if start < 0 {
				start = 0
			}
			end = start + listHeight
			if end > len(m.filteredItems) {
				end = len(m.filteredItems)
				start = end - listHeight
				if start < 0 {
					start = 0
				}
			}
		}

		var listContent strings.Builder
		for i := start; i < end; i++ {
			listContent.WriteString(RenderItem(m.filteredItems[i], i, m.cursor, listWidth-2, m.frame, m.animations))
			if i < end-1 {
				listContent.WriteString("\n")
			}
		}

		// Keep the list container height stable even when only a few items are visible.
		b.WriteString(listStyle.Width(listWidth).Height(listHeight).Render(listContent.String()))
	}

	selectedCount := 0
	failedCount := 0
	for _, item := range m.allItems {
		if item.Selected {
			selectedCount++
		}
		if item.Failed {
			failedCount++
		}
	}

	var statusIcon string
	if selectedCount > 0 {
		statusIcon = "◆ "
		if m.animations {
			icons := []string{"◆", "◇", "◆", "◇"}
			statusIcon = icons[m.frame%len(icons)] + " "
		}
	}

	status := fmt.Sprintf("%s%d selected • %d failing • %d/%d shown", statusIcon, selectedCount, failedCount, len(m.filteredItems), len(m.allItems))
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(status))

	b.WriteString("\n")
	help := m.keyMap.ShortHelp(m.compactHelp)
	b.WriteString(helpStyle.Render(help))

	return appStyle.Render(b.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
