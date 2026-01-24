package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 2).
			MarginBottom(1)

	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	fileCountStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

	fileListStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			MarginLeft(2)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))
)

func runTestsCmd(files []string) tea.Cmd {
	return func() tea.Msg {
		return runTestsMsg{files: files}
	}
}

type runTestsMsg struct {
	files []string
}

func PrintRunBanner(files []string) {
	logo := `
  ███████╗███████╗████████╗
  ██╔════╝╚══███╔╝╚══██╔══╝
  █████╗    ███╔╝    ██║   
  ██╔══╝   ███╔╝     ██║   
  ███████╗███████╗   ██║   
  ╚══════╝╚══════╝   ╚═╝   `

	fmt.Println(logoStyle.Render(logo))
	fmt.Println()

	header := fmt.Sprintf("  Running %s test file(s)", fileCountStyle.Render(fmt.Sprintf("%d", len(files))))
	fmt.Println(bannerStyle.Render(header))

	if len(files) <= 10 {
		for _, f := range files {
			fmt.Println(fileListStyle.Render("• " + f))
		}
	} else {
		for _, f := range files[:5] {
			fmt.Println(fileListStyle.Render("• " + f))
		}
		fmt.Println(fileListStyle.Render(fmt.Sprintf("  ... and %d more ...", len(files)-8)))
		for _, f := range files[len(files)-3:] {
			fmt.Println(fileListStyle.Render("• " + f))
		}
	}

	fmt.Println()
	divider := strings.Repeat("─", 50)
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
}

func ExecuteMixTest(files []string) error {
	if len(files) == 0 {
		return nil
	}

	PrintRunBanner(files)

	mixPath, err := exec.LookPath("mix")
	if err != nil {
		return err
	}

	args := make([]string, 0, len(files)+2)
	args = append(args, "mix", "test")
	args = append(args, files...)

	return syscall.Exec(mixPath, args, os.Environ())
}
