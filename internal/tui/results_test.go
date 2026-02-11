package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseRerunAction(t *testing.T) {
	if action, ok := parseRerunAction("r", false); !ok || action != RerunActionAll {
		t.Fatalf("expected r to map to rerun all, got action=%v ok=%v", action, ok)
	}

	if action, ok := parseRerunAction("rf", true); !ok || action != RerunActionFailed {
		t.Fatalf("expected rf to map to rerun failed, got action=%v ok=%v", action, ok)
	}

	if _, ok := parseRerunAction("rf", false); ok {
		t.Fatalf("expected rf to be rejected when failed rerun is unavailable")
	}

	if action, ok := parseRerunAction("q", true); !ok || action != RerunActionQuit {
		t.Fatalf("expected q to map to quit, got action=%v ok=%v", action, ok)
	}
}

func TestRenderResultsScreenIncludesSummaryAndFailures(t *testing.T) {
	ApplyTheme("default")

	outcome := TestRunOutcome{
		Stats: TestRunStats{
			Tests:         8,
			DescribeCount: 3,
			Failures:      1,
			Duration:      "0.12 seconds",
		},
		Failures: []FailureDetail{
			{
				Index:   1,
				Name:    "test signs in user",
				Module:  "MyApp.AuthTest",
				File:    "test/auth_test.exs",
				Line:    44,
				Details: "Assertion with == failed",
			},
		},
	}

	rendered := RenderResultsScreen(outcome, 2)
	expected := []string{
		"EZTest Failures",
		"failing test(s)",
		"Failed Tests",
		"test signs in user",
		"Module: MyApp.AuthTest",
		"test/auth_test.exs:44",
	}
	for _, snippet := range expected {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected rendered results to contain %q; output was:\n%s", snippet, rendered)
		}
	}

	if strings.Contains(rendered, "Status: PASS") {
		t.Fatalf("failure screen should not include pass status, got:\n%s", rendered)
	}
}

func TestCompactErrorDetailsRemovesStackNoise(t *testing.T) {
	raw := `
Assertion with == failed
code: assert user.id == nil
left: 123
right: nil
stacktrace:
  test/user_test.exs:33: (test)
  (elixir 1.17.0) lib/enum.ex:123: Enum.map/2
`

	compact := compactErrorDetails(raw)
	if strings.Contains(strings.ToLower(compact), "stacktrace") {
		t.Fatalf("expected stacktrace label to be removed, got:\n%s", compact)
	}
	if strings.Contains(compact, "Enum.map/2") {
		t.Fatalf("expected internal stack frame to be removed, got:\n%s", compact)
	}
	for _, expected := range []string{
		"Assertion with == failed",
		"Code: assert user.id == nil",
		"Left: 123",
		"Right: nil",
	} {
		if !strings.Contains(compact, expected) {
			t.Fatalf("expected compact details to include %q, got:\n%s", expected, compact)
		}
	}
}

func TestExtractFallbackErrorFiltersTaskNoise(t *testing.T) {
	raw := `
.............................................................nonode@nohost 20:17:26.907 pid=<0.4050.0> [error] Task #PID<0.4050.0> started from #PID<0.960.0> terminating
terminating
** (KeyError) key :assessment_version not found in: nil
.......................................nonode@nohost 20:17:27.691 pid=<0.6704.0> [error] Task #PID<0.6704.0> started from #PID<0.4350.0> terminating
** (Ecto.NoResultsError) expected at least one result but got none in query:
`

	got := extractFallbackError(raw)
	if strings.Contains(strings.ToLower(got), "nonode@nohost") {
		t.Fatalf("expected task logger noise to be filtered out, got:\n%s", got)
	}
	if strings.Contains(strings.ToLower(got), "task #pid") {
		t.Fatalf("expected task pid lines to be filtered out, got:\n%s", got)
	}
	for _, expected := range []string{
		"** (KeyError) key :assessment_version not found in: nil",
		"** (Ecto.NoResultsError) expected at least one result but got none in query:",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected fallback details to include %q, got:\n%s", expected, got)
		}
	}
}

func TestRenderFailureResultsUsesRuntimeLabelWhenNoTestFailures(t *testing.T) {
	ApplyTheme("default")

	outcome := TestRunOutcome{
		Stats: TestRunStats{
			Failures: 0,
		},
		Failures:  []FailureDetail{},
		RawOutput: "** (KeyError) boom",
	}

	rendered := RenderResultsScreen(outcome, 1)
	if !strings.Contains(rendered, "runtime error(s)") {
		t.Fatalf("expected runtime error label when test failures=0, got:\n%s", rendered)
	}
	if strings.Contains(rendered, "failing test(s)") {
		t.Fatalf("did not expect failing test label when test failures=0, got:\n%s", rendered)
	}
}

func TestRenderRunGraphsIncludesLegendAndRate(t *testing.T) {
	ApplyTheme("default")

	outcome := TestRunOutcome{
		Stats: TestRunStats{
			Tests:         10,
			DescribeCount: 3,
			Failures:      2,
		},
	}

	rendered := renderRunGraphs(outcome)
	for _, snippet := range []string{
		"Run Graphs",
		"Pass:",
		"Fail:",
		"8/10",
		"2/10",
		"█",
		"░",
		"Pass rate: 80%",
		"Total tests: 10",
		"Describe blocks: 3",
	} {
		if !strings.Contains(rendered, snippet) {
			t.Fatalf("expected graph block to contain %q, got:\n%s", snippet, rendered)
		}
	}
}

func TestPromptRerunActionOffersRfWhenFailureStatsPresent(t *testing.T) {
	ApplyTheme("default")

	outcome := TestRunOutcome{
		Stats: TestRunStats{
			Failures: 1,
		},
	}

	var out bytes.Buffer
	action := PromptRerunAction(outcome, strings.NewReader("rf\n"), &out)
	if action != RerunActionFailed {
		t.Fatalf("expected rf action when failure stats are present, got %v", action)
	}
}
