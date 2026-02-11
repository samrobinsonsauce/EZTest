package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type RerunAction int

const (
	failureCardMaxWidth     = 96
	failureCardContentWidth = 90
)

const (
	RerunActionQuit RerunAction = iota
	RerunActionAll
	RerunActionFailed
)

func PrintResultsScreen(outcome TestRunOutcome, exitCode int) {
	fmt.Println(RenderResultsScreen(outcome, exitCode))
}

func RenderResultsScreen(outcome TestRunOutcome, exitCode int) string {
	if exitCode != 0 {
		return renderFailureResults(outcome)
	}

	return renderPassingResults(outcome)
}

func renderPassingResults(outcome TestRunOutcome) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("EZTest Results"))
	b.WriteString("\n")
	b.WriteString(statusStyle.Render("Status: PASS"))
	b.WriteString("\n\n")

	b.WriteString(listStyle.Render(strings.Join([]string{
		fmt.Sprintf("Tests: %s", formatMetric(outcome.Stats.Tests)),
		fmt.Sprintf("Describe blocks: %d", outcome.Stats.DescribeCount),
		fmt.Sprintf("Failures: %s", formatMetric(outcome.Stats.Failures)),
		fmt.Sprintf("Duration: %s", formatDuration(outcome.Stats.Duration)),
	}, "\n")))

	b.WriteString("\n\n")
	b.WriteString(renderRunGraphs(outcome))
	b.WriteString("\n\n")
	b.WriteString(statusStyle.Render("All tests passed."))

	return b.String()
}

func renderFailureResults(outcome TestRunOutcome) string {
	var b strings.Builder

	failures := outcome.Failures
	reportedTestFailures := 0
	if outcome.Stats.Failures >= 0 {
		reportedTestFailures = outcome.Stats.Failures
	} else {
		reportedTestFailures = len(failures)
	}

	headerLabel := "failing test(s)"
	if reportedTestFailures == 0 {
		headerLabel = "runtime error(s)"
	}
	if len(failures) == 0 {
		// Keep fail mode error-only, even when detailed parse fails.
		fallback := FailureDetail{
			Index:   0,
			Name:    "Runtime error",
			Module:  "",
			File:    "",
			Line:    0,
			Details: extractFallbackError(outcome.RawOutput),
		}
		failures = []FailureDetail{fallback}
	}

	b.WriteString(errorStyle.Render("EZTest Failures"))
	b.WriteString("\n")
	b.WriteString(errorStyle.Render(fmt.Sprintf("%d %s", max(reportedTestFailures, len(failures)), headerLabel)))
	b.WriteString("\n\n")
	b.WriteString(renderRunGraphs(outcome))
	b.WriteString("\n\n")

	sectionHeader := "Failed Tests"
	if reportedTestFailures == 0 {
		sectionHeader = "Error Details"
	}
	sectionStyle := lipgloss.NewStyle().
		Foreground(errorStyle.GetForeground()).
		Bold(true).
		Underline(true)
	b.WriteString(sectionStyle.Render(sectionHeader))
	b.WriteString("\n\n")

	for i, failure := range failures {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(renderFailureCard(failure))
	}

	return b.String()
}

func renderFailureCard(failure FailureDetail) string {
	titleLine := fmt.Sprintf("✗ #%d  %s", failure.Index, failure.Name)
	if failure.Index <= 0 {
		titleLine = "✗ " + failure.Name
	}
	if strings.TrimSpace(failure.Name) == "" {
		if failure.Index > 0 {
			titleLine = fmt.Sprintf("✗ #%d  Test failure", failure.Index)
		} else {
			titleLine = "✗ Runtime error"
		}
	}

	meta := make([]string, 0, 2)
	if failure.Module != "" {
		meta = append(meta, "Module: "+failure.Module)
	}
	location := failure.File
	if failure.Line > 0 {
		location = fmt.Sprintf("%s:%d", failure.File, failure.Line)
	}
	if location != "" {
		meta = append(meta, "File: "+location)
	}

	detail := compactErrorDetails(failure.Details)
	if detail == "" {
		detail = "No parsed error detail available."
	}
	detail = renderFailureDetails(detail)

	wrapStyle := lipgloss.NewStyle().MaxWidth(failureCardContentWidth)
	moduleStyle := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	fileStyle := lipgloss.NewStyle().Foreground(dimTextColor)
	titleStyle := errorStyle.Bold(true)

	lines := []string{titleStyle.Render(wrapStyle.Render(titleLine))}
	for _, metaLine := range meta {
		styledLine := moduleStyle.Render(wrapStyle.Render(metaLine))
		if strings.HasPrefix(metaLine, "File: ") {
			styledLine = fileStyle.Render(wrapStyle.Render(metaLine))
		}
		lines = append(lines, styledLine)
	}
	lines = append(lines, wrapStyle.Render(detail))

	card := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(errorStyle.GetForeground()).
		Padding(0, 1).
		MaxWidth(failureCardMaxWidth).
		Render(card)
}

func renderFailureDetails(detail string) string {
	lines := strings.Split(detail, "\n")
	rendered := make([]string, 0, len(lines))

	for _, line := range lines {
		rendered = append(rendered, styleFailureDetailLine(line))
	}

	return strings.Join(rendered, "\n")
}

func styleFailureDetailLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)
	errTone := lipgloss.NewStyle().Foreground(errorStyle.GetForeground()).Bold(true)
	codeLabel := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	leftLabel := lipgloss.NewStyle().Foreground(errorStyle.GetForeground()).Bold(true)
	rightLabel := lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
	bodyTone := lipgloss.NewStyle().Foreground(textColor)
	strongBody := lipgloss.NewStyle().Foreground(textColor).Bold(true)

	switch {
	case strings.HasPrefix(trimmed, "** ("):
		return errTone.Render(trimmed)
	case strings.HasPrefix(lower, "code:"):
		return codeLabel.Render("Code:") + " " + bodyTone.Render(strings.TrimSpace(trimmed[5:]))
	case strings.HasPrefix(lower, "left:"):
		return leftLabel.Render("Left:") + " " + errTone.UnsetBold().Render(strings.TrimSpace(trimmed[5:]))
	case strings.HasPrefix(lower, "right:"):
		return rightLabel.Render("Right:") + " " + rightLabel.UnsetBold().Render(strings.TrimSpace(trimmed[6:]))
	case strings.HasPrefix(lower, "expected:"):
		return rightLabel.Render("Expected:") + " " + bodyTone.Render(strings.TrimSpace(trimmed[len("expected:"):]))
	case strings.HasPrefix(lower, "actual:"):
		return leftLabel.Render("Actual:") + " " + bodyTone.Render(strings.TrimSpace(trimmed[len("actual:"):]))
	case strings.Contains(lower, "assertion"):
		return strongBody.Render(trimmed)
	default:
		return bodyTone.Render(trimmed)
	}
}

func renderRunGraphs(outcome TestRunOutcome) string {
	total, pass, fail := graphCounts(outcome)
	rowTotal := max(total, pass+fail)

	lines := []string{
		titleStyle.Render("Run Graphs"),
		renderRunGraphRow("Pass", pass, rowTotal),
		renderRunGraphRow("Fail", fail, rowTotal),
	}

	if total > 0 {
		lines = append(lines, helpStyle.UnsetMargins().Render(
			fmt.Sprintf("Pass rate: %d%%  •  Total tests: %d", safePercent(pass, total), total),
		))
	} else {
		lines = append(lines, helpStyle.UnsetMargins().Render(
			"Pass rate: unknown  •  Total tests: unknown",
		))
	}
	lines = append(lines, helpStyle.UnsetMargins().Render(
		fmt.Sprintf("Describe blocks: %d", max(outcome.Stats.DescribeCount, 0)),
	))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

func renderRunGraphRow(label string, value, total int) string {
	labelStyle := lipgloss.NewStyle().Foreground(dimTextColor).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(textColor).Bold(true)

	if total <= 0 {
		total = max(value, 1)
	}
	value = min(max(value, 0), total)

	labelCell := lipgloss.NewStyle().Width(10).Render(label + ":")
	valueCell := lipgloss.NewStyle().Width(10).Align(lipgloss.Right).Render(fmt.Sprintf("%d/%d", value, total))

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(labelCell),
		" ",
		renderCanonicalProgressBar(value, total, 0),
		"  ",
		valueStyle.Render(valueCell),
	)
}

func graphCounts(outcome TestRunOutcome) (total, pass, fail int) {
	fail = outcome.Stats.Failures
	if fail < 0 {
		fail = len(outcome.Failures)
	}
	if fail < 0 {
		fail = 0
	}

	total = outcome.Stats.Tests
	if total < 0 {
		total = fail
	}
	if total < fail {
		total = fail
	}

	pass = total - fail
	if pass < 0 {
		pass = 0
	}

	return total, pass, fail
}

func safePercent(value, total int) int {
	if total <= 0 {
		return 0
	}
	if value < 0 {
		value = 0
	}
	if value > total {
		value = total
	}
	return int(float64(value) / float64(total) * 100)
}

func compactErrorDetails(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	out := make([]string, 0, 8)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "stacktrace:") {
			break
		}
		if isStackFrameLine(trimmed) {
			continue
		}
		if strings.HasPrefix(lower, "the following output was logged:") {
			continue
		}

		out = append(out, formatErrorLine(trimmed))
		if len(out) >= 8 {
			break
		}
	}

	return strings.Join(out, "\n")
}

func isStackFrameLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// Typical ExUnit stack frame line: test/foo_test.exs:12: ...
	if strings.Contains(line, "_test.exs:") {
		return true
	}
	// Erlang/Elixir internal frames often look like "(elixir 1.x) lib/..."
	if strings.HasPrefix(line, "(") && strings.Contains(line, ") ") {
		return true
	}
	return false
}

func formatErrorLine(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.HasPrefix(lower, "code:"):
		return "Code: " + strings.TrimSpace(line[5:])
	case strings.HasPrefix(lower, "left:"):
		return "Left: " + strings.TrimSpace(line[5:])
	case strings.HasPrefix(lower, "right:"):
		return "Right: " + strings.TrimSpace(line[6:])
	default:
		return line
	}
}

func extractFallbackError(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "No error output captured."
	}

	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	out := make([]string, 0, 6)
	seen := make(map[string]struct{})
	for _, line := range lines {
		trimmed := normalizeRuntimeLine(ansiEscapePattern.ReplaceAllString(line, ""))
		if trimmed == "" {
			continue
		}
		if !isRelevantRuntimeErrorLine(trimmed) {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
		if len(out) >= 6 {
			break
		}
	}
	if len(out) == 0 {
		return "Test run failed and no structured error lines were found."
	}
	return strings.Join(out, "\n")
}

func normalizeRuntimeLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, ".")
	line = strings.TrimSpace(line)
	return line
}

func isRelevantRuntimeErrorLine(line string) bool {
	lower := strings.ToLower(line)

	if line == "" {
		return false
	}
	if strings.Contains(lower, "nonode@nohost") && strings.Contains(lower, "[error]") {
		return false
	}
	if strings.Contains(lower, "task #pid<") && strings.Contains(lower, "terminating") {
		return false
	}
	if lower == "terminating" {
		return false
	}

	if strings.HasPrefix(line, "** (") || strings.HasPrefix(line, "** ") {
		return true
	}
	if strings.Contains(lower, "compilation error") || strings.Contains(lower, "mix ") {
		return true
	}
	if strings.Contains(lower, "error") || strings.Contains(lower, "failure") || strings.HasPrefix(lower, "exit:") {
		return true
	}

	return false
}

func PromptRerunAction(outcome TestRunOutcome, in io.Reader, out io.Writer) RerunAction {
	reader := bufio.NewReader(in)
	allowRerunFailed := len(outcome.FailedFiles) > 0 ||
		len(outcome.Failures) > 0 ||
		outcome.Stats.Failures > 0

	for {
		if allowRerunFailed {
			fmt.Fprint(out, "Action [r] rerun, [rf] rerun failed, [q] quit: ")
		} else {
			fmt.Fprint(out, "Action [r] rerun, [q] quit: ")
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return RerunActionQuit
		}

		if action, ok := parseRerunAction(input, allowRerunFailed); ok {
			return action
		}
	}
}

func parseRerunAction(input string, allowRerunFailed bool) (RerunAction, bool) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch normalized {
	case "q", "quit", "":
		return RerunActionQuit, true
	case "r":
		return RerunActionAll, true
	case "rf":
		if allowRerunFailed {
			return RerunActionFailed, true
		}
		return RerunActionQuit, false
	default:
		return RerunActionQuit, false
	}
}

func formatMetric(value int) string {
	if value < 0 {
		return "unknown"
	}
	return fmt.Sprintf("%d", value)
}

func formatDuration(duration string) string {
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return "unknown"
	}
	return duration
}
