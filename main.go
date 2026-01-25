package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samrobinsonsauce/eztest/internal/config"
	"github.com/samrobinsonsauce/eztest/internal/testfile"
	"github.com/samrobinsonsauce/eztest/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help")
	runDirect := flag.Bool("r", false, "Run saved tests directly without opening TUI")
	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("ezt version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get current directory: %v\n", err)
		os.Exit(1)
	}

	if *runDirect {
		selections, err := config.GetProjectSelections(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading saved tests: %v\n", err)
			os.Exit(1)
		}
		if len(selections) == 0 {
			fmt.Fprintf(os.Stderr, "No tests saved. Run 'ezt' first to select tests.\n")
			os.Exit(1)
		}
		if err := tui.ExecuteMixTest(selections); err != nil {
			fmt.Fprintf(os.Stderr, "Error running mix test: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	testFiles, err := testfile.FindTestFiles(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	selections, err := config.GetProjectSelections(cwd)
	if err != nil {
		selections = []string{}
	}

	model := tui.NewModel(testFiles, cwd, selections)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	m, ok := finalModel.(tui.Model)
	if !ok {
		os.Exit(1)
	}

	if m.IsQuitting() {
		os.Exit(0)
	}

	files := m.GetFilesToRun()
	if len(files) == 0 {
		fmt.Println("No tests selected.")
		os.Exit(0)
	}

	if err := tui.ExecuteMixTest(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error running mix test: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	help := `ezt - Elixir Test Selector

A TUI for selecting and running Elixir tests.

USAGE:
    ezt [OPTIONS]

OPTIONS:
    -r           Run saved tests directly (skip TUI)
    --help       Show this help message
    --version    Show version information

KEYBINDINGS:
    ↑ / Ctrl+k   Move cursor up
    ↓ / Ctrl+j   Move cursor down
    Tab          Toggle selection on current item
    Ctrl+a       Select all visible (filtered) items
    Ctrl+d       Deselect all items
    Enter        Run selected tests with mix test
    Ctrl+s       Save selections and quit (without running)
    Esc          Quit without saving

EXAMPLES:
    ezt          Open TUI to select and run tests
    ezt -r       Run previously saved tests directly

USAGE:
    Navigate to your Elixir/Phoenix project and run 'ezt'.
    Type to filter tests, use Tab to select, Enter to run.

    Selections are saved per-project in ~/.config/ezt/state.json
`
	fmt.Print(help)
}
