package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samrobinsonsauce/eztest/internal/config"
	"github.com/samrobinsonsauce/eztest/internal/testfile"
	"github.com/samrobinsonsauce/eztest/internal/tui"
)

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	executeMixTest = tui.ExecuteMixTest
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help")
	runDirect := flag.Bool("r", false, "Run saved tests directly without opening TUI")
	runFailed := flag.Bool("f", false, "Run last failed tests directly without opening TUI")
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

	appSettings, err := config.LoadAppSettings()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		appSettings = config.DefaultAppSettings()
	}
	tui.ApplyTheme(appSettings.Theme)

	if *runDirect && *runFailed {
		fmt.Fprintf(os.Stderr, "Use either -r or -f, not both.\n")
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
		os.Exit(runAndPersistFailures(cwd, selections))
	}

	if *runFailed {
		failures, err := config.GetProjectFailures(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading failed tests: %v\n", err)
			os.Exit(1)
		}
		if len(failures) == 0 {
			fmt.Fprintf(os.Stderr, "No failed tests saved. Run tests first to capture failures.\n")
			os.Exit(1)
		}
		os.Exit(runAndPersistFailures(cwd, failures))
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

	model := tui.NewModel(
		testFiles,
		cwd,
		selections,
		failuresForProject(cwd),
		tui.NewKeyMap(appSettings.Keybinds),
		appSettings.UI,
	)
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

	os.Exit(runAndPersistFailures(cwd, files))
}

func printHelp() {
	configPath, err := config.GetAppConfigPath()
	if err != nil {
		configPath = "~/.config/eztest/config.json"
	}

	help := fmt.Sprintf(`ezt - Elixir Test Selector

A TUI for selecting and running Elixir tests.

USAGE:
    ezt [OPTIONS]

OPTIONS:
    -r           Run saved tests directly (skip TUI)
    -f           Run last failed tests directly (skip TUI)
    --help       Show this help message
    --version    Show version information

KEYBINDINGS:
    (defaults shown below; configurable in %s)
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
    ezt -f       Run previously failed tests directly

USAGE:
    Navigate to your Elixir/Phoenix project and run 'ezt'.
    Type to filter tests, use Tab to select, Enter to run.

    Selections are saved per-project in ~/.config/eztest/state.json
`, configPath)
	fmt.Print(help)
}

func failuresForProject(projectDir string) []string {
	failures, err := config.GetProjectFailures(projectDir)
	if err != nil {
		return []string{}
	}
	return failures
}

func runAndPersistFailures(projectDir string, files []string) int {
	outcome, err := executeMixTest(files)

	var exitErr *exec.ExitError
	if err == nil || errors.As(err, &exitErr) {
		if saveErr := config.SaveProjectFailures(projectDir, outcome.FailedFiles); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to persist failed tests: %v\n", saveErr)
		}
	}

	if err == nil {
		return 0
	}

	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	fmt.Fprintf(os.Stderr, "Error running mix test: %v\n", err)
	return 1
}
