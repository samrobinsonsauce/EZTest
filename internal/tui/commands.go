package tui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type TestRunStats struct {
	Tests         int
	DescribeCount int
	Failures      int
	Duration      string
}

type FailureDetail struct {
	Index   int
	Name    string
	Module  string
	File    string
	Line    int
	Details string
}

type TestRunOutcome struct {
	FailedFiles []string
	Stats       TestRunStats
	Failures    []FailureDetail
	RawOutput   string
}

type RunOptions struct {
	FailFast bool
}

var (
	ansiEscapePattern    = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	describePattern      = regexp.MustCompile(`(?m)^\s*describe\s+["']`)
	testSummaryPattern   = regexp.MustCompile(`(?m)(\d+)\s+tests?,\s+(\d+)\s+failures?`)
	durationPattern      = regexp.MustCompile(`(?m)^Finished in ([^\n]+)$`)
	failureHeaderPattern = regexp.MustCompile(`^\s*(\d+)\)\s+(.+)\s+\(([^)]+)\)\s*$`)
	failureFilePattern   = regexp.MustCompile(`^\s*([^\s]+_test\.exs):(\d+)`)
)

type fileTestCountCacheEntry struct {
	Size       int64
	ModTimeNix int64
	Count      int
}

var testCountCache sync.Map
var (
	runOptionsMu      sync.RWMutex
	currentRunOptions RunOptions
)

func SetRunOptions(options RunOptions) {
	runOptionsMu.Lock()
	currentRunOptions = options
	runOptionsMu.Unlock()
}

func getRunOptions() RunOptions {
	runOptionsMu.RLock()
	defer runOptionsMu.RUnlock()
	return currentRunOptions
}

func PrintRunBanner(files []string, failedOnly bool) {
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
	if failedOnly {
		header = "  Running previously failed tests (--failed)"
	}
	fmt.Println(bannerStyle.Render(header))

	if !failedOnly {
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
	} else {
		fmt.Println(fileListStyle.Render("• mix test --failed"))
		fmt.Println(fileListStyle.Render("• Uses ExUnit's previous-failure tracking"))
	}

	fmt.Println()
	divider := strings.Repeat("─", 50)
	fmt.Println(dividerStyle.Render(divider))
	fmt.Println()
}

func ExecuteMixTest(files []string) (TestRunOutcome, error) {
	return executeMixTest(files, false)
}

func ExecuteMixTestFailed() (TestRunOutcome, error) {
	return executeMixTest(nil, true)
}

func executeMixTest(files []string, failedOnly bool) (TestRunOutcome, error) {
	outcome := TestRunOutcome{
		FailedFiles: []string{},
		Stats: TestRunStats{
			Tests:         -1,
			DescribeCount: countDescribeBlocks(files),
			Failures:      -1,
		},
		Failures:  []FailureDetail{},
		RawOutput: "",
	}
	if !failedOnly && len(files) == 0 {
		return outcome, nil
	}

	PrintRunBanner(files, failedOnly)

	scannedTestCount := 0
	if !failedOnly {
		scannedTestCount = countSelectedTests(files)
	}
	scanInfoStyle := lipgloss.NewStyle().Foreground(dimTextColor)
	scanOkStyle := lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
	if failedOnly {
		fmt.Println(scanInfoStyle.Render("Scan: skipped for --failed run."))
	} else if scannedTestCount > 0 {
		fmt.Println(scanOkStyle.Render(fmt.Sprintf("Scan: %d tests discovered in selected files.", scannedTestCount)))
	} else {
		fmt.Println(scanInfoStyle.Render("Scan: no test macros discovered; tracking live results only."))
	}
	fmt.Println()

	mixPath, err := exec.LookPath("mix")
	if err != nil {
		return outcome, err
	}

	args := make([]string, 0, len(files)+2)
	args = buildMixTestArgs(files, getRunOptions(), failedOnly)

	cmd := exec.Command(mixPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	var output bytes.Buffer
	var outputMu sync.Mutex

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return outcome, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return outcome, err
	}

	progress := newProgressTracker(canRenderLiveProgress(), scannedTestCount)
	if progress.enabled {
		progress.RenderInitial()
	}

	if err := cmd.Start(); err != nil {
		return outcome, err
	}

	var wg sync.WaitGroup
	readErrs := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		readErrs <- consumeOutput(stdoutPipe, &output, &outputMu, progress, true)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		readErrs <- consumeOutput(stderrPipe, &output, &outputMu, progress, false)
	}()

	wg.Wait()
	waitErr := cmd.Wait()
	close(readErrs)

	for readErr := range readErrs {
		if readErr != nil {
			err = readErr
			break
		}
	}
	if err == nil {
		err = waitErr
	}

	outputMu.Lock()
	outcome.RawOutput = output.String()
	outputMu.Unlock()
	outcome.Stats = parseTestRunStats(outcome.RawOutput, files)
	outcome.Failures = parseFailureDetails(outcome.RawOutput)
	if failedOnly {
		outcome.FailedFiles = failedFilesFromDetails(outcome.Failures)
		if len(outcome.FailedFiles) == 0 {
			outcome.FailedFiles = extractFailedFiles(outcome.RawOutput, nil)
		}
	} else {
		outcome.FailedFiles = extractFailedFiles(outcome.RawOutput, files)
		if len(outcome.FailedFiles) == 0 {
			outcome.FailedFiles = failedFilesFromDetails(outcome.Failures)
		}
	}

	if outcome.Stats.Failures < 0 {
		outcome.Stats.Failures = len(outcome.Failures)
	}
	progress.SyncWithSummary(outcome.Stats)
	progress.Finish()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && !failedOnly && len(outcome.FailedFiles) == 0 {
			outcome.FailedFiles = uniqueSortedFiles(files)
		}
		return outcome, err
	}

	return outcome, nil
}

func consumeOutput(
	reader io.Reader,
	output *bytes.Buffer,
	outputMu *sync.Mutex,
	progress *progressTracker,
	trackProgress bool,
) error {
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunk := buf[:n]

			outputMu.Lock()
			_, _ = output.Write(chunk)
			outputMu.Unlock()

			if trackProgress {
				progress.Consume(chunk)
			}
		}

		if err == io.EOF {
			return nil
		}
		if errors.Is(err, os.ErrClosed) {
			// Pipe can close as the process exits; that's not a user-facing failure.
			return nil
		}
		if err != nil {
			return err
		}
	}
}

type progressTracker struct {
	enabled bool
	total   int

	passCount  int
	failCount  int
	otherCount int

	atLineStart     bool
	lineIsProgress  bool
	hasPendingFirst bool
	pendingFirst    byte

	renderedLen int
	lastRender  time.Time
	frame       int
}

func newProgressTracker(enabled bool, total int) *progressTracker {
	return &progressTracker{
		enabled:        enabled,
		total:          total,
		atLineStart:    true,
		lineIsProgress: true,
	}
}

func (p *progressTracker) RenderInitial() {
	p.render(true)
}

func (p *progressTracker) Consume(chunk []byte) {
	for _, b := range chunk {
		p.consumeByte(b)
	}
}

func (p *progressTracker) consumeByte(b byte) {
	switch b {
	case '\n', '\r':
		p.hasPendingFirst = false
		p.atLineStart = true
		p.lineIsProgress = true
		return
	}

	if p.atLineStart {
		if b == ' ' || b == '\t' {
			return
		}
		p.atLineStart = false
		p.lineIsProgress = isProgressRune(b)
		if p.lineIsProgress {
			p.pendingFirst = b
			p.hasPendingFirst = true
			p.commitProgressRune(b)
		}
		return
	}

	if !p.lineIsProgress {
		return
	}

	if p.hasPendingFirst {
		if !isProgressRune(b) {
			p.rollbackProgressRune(p.pendingFirst)
			p.hasPendingFirst = false
			p.lineIsProgress = false
			return
		}
		p.hasPendingFirst = false
	}

	if !isProgressRune(b) {
		p.lineIsProgress = false
		return
	}
	p.commitProgressRune(b)
}

func (p *progressTracker) rollbackProgressRune(b byte) {
	switch b {
	case '.':
		p.passCount = max(0, p.passCount-1)
	case 'F':
		p.failCount = max(0, p.failCount-1)
	case '*', '?', 'S':
		p.otherCount = max(0, p.otherCount-1)
	}
}

func (p *progressTracker) commitProgressRune(b byte) {
	switch b {
	case '.':
		p.passCount++
	case 'F':
		p.failCount++
	case '*', '?', 'S':
		p.otherCount++
	default:
		return
	}
	p.frame++
	p.render(false)
}

func (p *progressTracker) Finish() {
	if !p.enabled {
		return
	}
	if p.totalSeen() == 0 {
		return
	}
	p.render(true)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout)
}

func (p *progressTracker) Snapshot() (pass, fail, other int) {
	return p.passCount, p.failCount, p.otherCount
}

func (p *progressTracker) SyncWithSummary(stats TestRunStats) {
	if stats.Failures < 0 {
		return
	}
	p.failCount = stats.Failures
}

func (p *progressTracker) render(force bool) {
	if !p.enabled {
		return
	}

	now := time.Now()
	if !force && now.Sub(p.lastRender) < 75*time.Millisecond {
		return
	}

	line := p.buildLine()

	padding := ""
	if p.renderedLen > len(line) {
		padding = strings.Repeat(" ", p.renderedLen-len(line))
	}

	fmt.Fprintf(os.Stdout, "\r%s%s", line, padding)
	p.renderedLen = len(line)
	p.lastRender = now
}

func (p *progressTracker) buildLine() string {
	spinners := []string{"◐", "◓", "◑", "◒"}
	spin := spinners[p.frame%len(spinners)]

	seen := p.totalSeen()
	progressPrefix := fmt.Sprintf("%s  %s", spin, p.progressLabel(seen))
	metrics := fmt.Sprintf("  ✓ %d   ✗ %d", p.passCount, p.failCount)
	if p.otherCount > 0 {
		metrics += fmt.Sprintf("   ? %d", p.otherCount)
	}

	titleInline := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	metricInline := lipgloss.NewStyle().Foreground(textColor).Bold(true)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleInline.Render(progressPrefix),
		"  ",
		p.renderBar(seen),
		metricInline.Render(metrics),
	)
}

func (p *progressTracker) progressLabel(seen int) string {
	return formatProgressLabel(seen, p.total)
}

func (p *progressTracker) renderBar(seen int) string {
	return renderCanonicalProgressBar(seen, p.total, p.frame)
}

func (p *progressTracker) totalSeen() int {
	return p.passCount + p.failCount + p.otherCount
}

func isProgressRune(b byte) bool {
	switch b {
	case '.', 'F', '*', '?', 'S':
		return true
	default:
		return false
	}
}

func canRenderLiveProgress() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func buildMixTestArgs(files []string, options RunOptions, failedOnly bool) []string {
	// Keep suite execution going unless fail-fast is explicitly enabled.
	// This avoids project-level ExUnit max_failures settings aborting early.
	const nonFailFastMaxFailures = "2147483647"

	args := make([]string, 0, len(files)+5)
	args = append(args, "test")
	maxFailures := nonFailFastMaxFailures
	if options.FailFast {
		maxFailures = "1"
	}
	args = append(args, "--max-failures", maxFailures)
	if failedOnly {
		args = append(args, "--failed")
	}
	args = append(args, files...)
	return args
}

func parseTestRunStats(output string, runFiles []string) TestRunStats {
	stats := TestRunStats{
		Tests:         -1,
		DescribeCount: countDescribeBlocks(runFiles),
		Failures:      -1,
		Duration:      "",
	}

	clean := ansiEscapePattern.ReplaceAllString(output, "")
	if matches := durationPattern.FindStringSubmatch(clean); len(matches) == 2 {
		stats.Duration = strings.TrimSpace(matches[1])
	}

	if matches := testSummaryPattern.FindStringSubmatch(clean); len(matches) == 3 {
		tests, err := strconv.Atoi(matches[1])
		if err == nil {
			stats.Tests = tests
		}
		failures, err := strconv.Atoi(matches[2])
		if err == nil {
			stats.Failures = failures
		}
	}

	return stats
}

func countDescribeBlocks(files []string) int {
	seen := make(map[string]struct{}, len(files))
	total := 0

	for _, file := range files {
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		total += len(describePattern.FindAll(data, -1))
	}

	return total
}

func countSelectedTests(files []string) int {
	seen := make(map[string]struct{}, len(files))
	total := 0

	for _, file := range files {
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}

		count, err := countTestsInFile(file)
		if err != nil {
			continue
		}
		total += count
	}

	return total
}

func countTestsInFile(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if cached, ok := testCountCache.Load(path); ok {
		entry := cached.(fileTestCountCacheEntry)
		if entry.Size == info.Size() && entry.ModTimeNix == info.ModTime().UnixNano() {
			return entry.Count, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	count := 0
	for scanner.Scan() {
		line := bytes.TrimLeft(scanner.Bytes(), " \t")
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if isTestMacroLine(line) {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return count, err
	}

	testCountCache.Store(path, fileTestCountCacheEntry{
		Size:       info.Size(),
		ModTimeNix: info.ModTime().UnixNano(),
		Count:      count,
	})

	return count, nil
}

func isTestMacroLine(line []byte) bool {
	return bytes.HasPrefix(line, []byte("test ")) ||
		bytes.HasPrefix(line, []byte("test\t")) ||
		bytes.HasPrefix(line, []byte("test(")) ||
		bytes.HasPrefix(line, []byte("property ")) ||
		bytes.HasPrefix(line, []byte("property\t")) ||
		bytes.HasPrefix(line, []byte("property("))
}

func parseFailureDetails(output string) []FailureDetail {
	clean := ansiEscapePattern.ReplaceAllString(output, "")
	lines := strings.Split(clean, "\n")
	details := make([]FailureDetail, 0)

	for i := 0; i < len(lines); i++ {
		header := failureHeaderPattern.FindStringSubmatch(lines[i])
		if len(header) != 4 {
			continue
		}

		index, _ := strconv.Atoi(header[1])
		detail := FailureDetail{
			Index:   index,
			Name:    strings.TrimSpace(header[2]),
			Module:  strings.TrimSpace(header[3]),
			File:    "",
			Line:    0,
			Details: "",
		}

		bodyLines := make([]string, 0)
		j := i + 1
		for ; j < len(lines); j++ {
			if failureHeaderPattern.MatchString(lines[j]) {
				break
			}
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "Finished in ") {
				break
			}

			line := strings.TrimRight(lines[j], "\r")
			if detail.File == "" {
				if fileMatch := failureFilePattern.FindStringSubmatch(strings.TrimSpace(line)); len(fileMatch) == 3 {
					detail.File = fileMatch[1]
					lineNo, err := strconv.Atoi(fileMatch[2])
					if err == nil {
						detail.Line = lineNo
					}
					continue
				}
			}

			trimmed := strings.TrimSpace(line)
			if trimmed == "" && len(bodyLines) == 0 {
				continue
			}
			bodyLines = append(bodyLines, trimmed)
		}

		for len(bodyLines) > 0 && bodyLines[len(bodyLines)-1] == "" {
			bodyLines = bodyLines[:len(bodyLines)-1]
		}
		detail.Details = strings.Join(bodyLines, "\n")
		details = append(details, detail)

		i = j - 1
	}

	return details
}

func extractFailedFiles(output string, runFiles []string) []string {
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

		if len(runSet) == 0 {
			found[path] = struct{}{}
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

func failedFilesFromDetails(details []FailureDetail) []string {
	found := make(map[string]struct{}, len(details))
	for _, detail := range details {
		path := strings.TrimSpace(detail.File)
		if path == "" {
			continue
		}
		found[path] = struct{}{}
	}

	files := make([]string, 0, len(found))
	for file := range found {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
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
