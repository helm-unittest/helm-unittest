package results

import (
	"bytes"
	"fmt"
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
