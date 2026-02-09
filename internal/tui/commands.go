package tui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func runTestsCmd(files []string) tea.Cmd {
	return func() tea.Msg {
		return runTestsMsg{files: files}
	}
}

type runTestsMsg struct {
	files []string
}

type TestRunOutcome struct {
	FailedFiles []string
}

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

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

func ExecuteMixTest(files []string) (TestRunOutcome, error) {
	outcome := TestRunOutcome{FailedFiles: []string{}}
	if len(files) == 0 {
		return outcome, nil
	}

	PrintRunBanner(files)

	mixPath, err := exec.LookPath("mix")
	if err != nil {
		return outcome, err
	}

	args := make([]string, 0, len(files)+1)
	args = append(args, "test")
	args = append(args, files...)

	cmd := exec.Command(mixPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = io.MultiWriter(os.Stderr, &output)

	err = cmd.Run()
	outcome.FailedFiles = extractFailedFiles(output.String(), files)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(outcome.FailedFiles) == 0 {
			outcome.FailedFiles = uniqueSortedFiles(files)
		}
		return outcome, err
	}

	return outcome, nil
}

func extractFailedFiles(output string, runFiles []string) []string {
	if len(runFiles) == 0 {
		return []string{}
	}

	output = ansiEscapePattern.ReplaceAllString(output, "")
	runSet := make(map[string]struct{}, len(runFiles))
	for _, file := range runFiles {
		runSet[file] = struct{}{}
	}

	found := make(map[string]struct{})
	for _, token := range strings.Fields(output) {
		path := pathFromToken(token)
		if path == "" {
			continue
		}

		if _, ok := runSet[path]; ok {
			found[path] = struct{}{}
			continue
		}

		for candidate := range runSet {
			if strings.HasSuffix(path, "/"+candidate) {
				found[candidate] = struct{}{}
			}
		}
	}

	failures := make([]string, 0, len(found))
	for file := range found {
		failures = append(failures, file)
	}
	sort.Strings(failures)
	return failures
}

func pathFromToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}

	token = strings.Trim(token, "\"'`()[]{}<>:,;")
	if token == "" {
		return ""
	}

	const suffix = "_test.exs"
	idx := strings.Index(token, suffix)
	if idx < 0 {
		return ""
	}

	path := token[:idx+len(suffix)]
	path = strings.TrimPrefix(path, "./")
	path = strings.ReplaceAll(path, "\\", "/")

	if path == "" {
		return ""
	}
	return path
}

func uniqueSortedFiles(files []string) []string {
	seen := make(map[string]struct{}, len(files))
	out := make([]string, 0, len(files))
	for _, file := range files {
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}
		out = append(out, file)
	}
	sort.Strings(out)
	return out
}
