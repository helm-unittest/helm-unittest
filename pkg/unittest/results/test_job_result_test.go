package results

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

// test print
func TestSkippedJob_PrintsSkippedMessage(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Skipped:     true,
	}

	tjr.print(pr, 1)
	assert.Contains(t, fmt.Sprintf("%s", pr.Writer), "- SKIPPED 'some job'")
}

func TestSkippedJob_NoPrintIfPassed(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      true,
		Skipped:     false,
	}

	tjr.print(pr, 1)
	assert.Empty(t, fmt.Sprintf("%s", pr.Writer))
}

// test Stringify
func TestStringify_NoErrorAndNoAssertions(t *testing.T) {
	tjr := TestJobResult{
		ExecError:     nil,
		AssertsResult: []*AssertionResult{},
	}
	expected := ""
	result := tjr.Stringify()
	assert.Equal(t, expected, result)
}

func TestStringify_WithExecError(t *testing.T) {
	tjr := TestJobResult{
		ExecError:     fmt.Errorf("execution error"),
		AssertsResult: []*AssertionResult{},
	}
	expected := "execution error\n"
	result := tjr.Stringify()
	assert.Equal(t, expected, result)
}

func TestStringify_WithAssertions(t *testing.T) {
	tjr := TestJobResult{
		ExecError: nil,
		AssertsResult: []*AssertionResult{
			{FailInfo: []string{"assertion error 1"}},
			{FailInfo: []string{"assertion error 2"}},
		},
	}
	expected := "\t\t - asserts[0] `` fail \n\t\t\t assertion error 1 \n"
	expected += "\t\t - asserts[0] `` fail \n\t\t\t assertion error 2 \n"
	result := tjr.Stringify()
	assert.Equal(t, expected, result)
}

func TestStringify_WithExecErrorAndAssertions(t *testing.T) {
	tjr := TestJobResult{
		ExecError: fmt.Errorf("execution error"),
		AssertsResult: []*AssertionResult{
			{FailInfo: []string{"assertion error 1"}},
			{FailInfo: []string{"assertion error 2"}},
		},
	}
	expected := "execution error\n"
	expected += "\t\t - asserts[0] `` fail \n\t\t\t assertion error 1 \n"
	expected += "\t\t - asserts[0] `` fail \n\t\t\t assertion error 2 \n"
	result := tjr.Stringify()
	assert.Equal(t, expected, result)
}

// test DebugOutput
func TestPrint_DebugOutputShownWhenJobFailed(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      false,
		DebugOutput: "#### file: chart/templates/deployment.yaml\nkind: Deployment\n",
		AssertsResult: []*AssertionResult{
			{Index: 0, FailInfo: []string{"expected", "actual"}, AssertType: "equal"},
		},
	}

	tjr.print(pr, 1)
	output := fmt.Sprintf("%s", pr.Writer)
	a := assert.New(t)
	a.Contains(output, "kind: Deployment", "debug output should accompany the failure")
	a.Contains(output, "asserts[0]", "the assertion failure must still be printed")
}

func TestPrint_DebugOutputShownWhenJobPassed(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      true,
		DebugOutput: "#### file: chart/templates/deployment.yaml\nkind: Deployment\n",
	}

	tjr.print(pr, 1)
	output := fmt.Sprintf("%s", pr.Writer)
	a := assert.New(t)
	// debug was explicitly asked for, so it must survive the passing-job early return
	a.Contains(output, "kind: Deployment", "debug output should print even when the job passed")
	a.NotContains(output, "asserts[", "a passing job has no assertion failures to report")
}

func TestPrint_NoDebugOutputWhenPassedAndNoDebug(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      true,
	}

	tjr.print(pr, 1)
	assert.Empty(t, fmt.Sprintf("%s", pr.Writer), "passing job without debug must stay silent")
}

func TestStringify_ExcludesDebugOutput(t *testing.T) {
	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      false,
		DebugOutput: "#### file: chart/templates/deployment.yaml\nkind: Deployment\n",
		AssertsResult: []*AssertionResult{
			{Index: 0, FailInfo: []string{"expected"}, AssertType: "equal"},
		},
	}

	// the XML formatters build their failure messages from Stringify, so rendered
	// manifests must not leak into junit/xunit/nunit/sonar reports
	a := assert.New(t)
	a.NotContains(tjr.Stringify(), "kind: Deployment")
	a.NotContains(tjr.StringifyToXmlAttribute(), "kind: Deployment")
	a.Contains(tjr.Stringify(), "equal")
}

func TestPrint_DebugOutputDoesNotRepeatJobNameWhenFailed(t *testing.T) {
	flag := false
	pr := printer.NewPrinter(new(bytes.Buffer), &flag)

	tjr := TestJobResult{
		DisplayName: "some job",
		Passed:      false,
		DebugOutput: "#### file: chart/templates/deployment.yaml\nkind: Deployment\n",
		AssertsResult: []*AssertionResult{
			{Index: 0, FailInfo: []string{"expected"}, AssertType: "equal"},
		},
	}

	tjr.print(pr, 1)
	output := fmt.Sprintf("%s", pr.Writer)
	assert.Equal(t, 1, strings.Count(output, "some job"),
		"the job name should be printed once, not repeated by the debug output")
}
