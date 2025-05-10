package results_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

func createTestSuiteResult(suiteDisplayName, testDisplayName, filePath, customInfo string,
	suiteError, testError error,
	failInfo []string,
	passed, inverse bool) TestSuiteResult {

	return TestSuiteResult{
		DisplayName: suiteDisplayName,
		FilePath:    filePath,
		Passed:      passed,
		ExecError:   suiteError,
		TestsResult: []*TestJobResult{
			{
				DisplayName: testDisplayName,
				Index:       0,
				Passed:      passed,
				ExecError:   testError,
				AssertsResult: []*AssertionResult{
					{
						Index:      0,
						FailInfo:   failInfo,
						Passed:     passed,
						AssertType: "equal",
						Not:        inverse,
						CustomInfo: customInfo,
					},
				},
				Duration: 0,
			},
		},
	}
}

func TestTestSuiteResultPrintPassed(t *testing.T) {
	buffer := new(bytes.Buffer)
	testPrinter := printer.NewPrinter(buffer, nil)
	given := createTestSuiteResult("A Test Suite", "A Test Case", "filePath", "", nil, nil, nil, true, false)

	expectedResult := fmt.Sprintf(" PASS  %s\t./%s\n", given.DisplayName, given.FilePath)
	given.Print(testPrinter, 0)

	a := assert.New(t)
	a.Equal(expectedResult, buffer.String())
}

func TestTestSuiteResultPrintFailed(t *testing.T) {
	buffer := new(bytes.Buffer)
	testPrinter := printer.NewPrinter(buffer, nil)
	given := createTestSuiteResult("A Test Suite", "A Test Case", "filePath", "", fmt.Errorf("An Error Occurred."), nil, nil, false, false)

	expectedResult := fmt.Sprintf(" FAIL  %s\t./%s\n\t- Execution Error: \n\t\t%s\n\n", given.DisplayName, given.FilePath, given.ExecError)
	given.Print(testPrinter, 0)

	a := assert.New(t)
	a.Equal(expectedResult, buffer.String())
}

func TestTestSuiteResultPrintFailedTestJob(t *testing.T) {
	buffer := new(bytes.Buffer)
	testPrinter := printer.NewPrinter(buffer, nil)
	given := createTestSuiteResult("A Test Suite", "A Test Case", "filePath", "", nil, fmt.Errorf("An Test Error."), nil, false, false)

	expectedResult := fmt.Sprintf(" FAIL  %s\t./%s\n\t- %s\n\t\tError: %s\n\n", given.DisplayName, given.FilePath, given.TestsResult[0].DisplayName, given.TestsResult[0].ExecError)
	given.Print(testPrinter, 0)

	a := assert.New(t)
	a.Equal(expectedResult, buffer.String())
}

func TestTestSuiteResultPrintFailedTestAsssertion(t *testing.T) {
	buffer := new(bytes.Buffer)
	testPrinter := printer.NewPrinter(buffer, nil)
	given := createTestSuiteResult("A Test Suite", "A Test Case", "filePath", "", nil, nil, []string{"Error1", "Error2"}, false, false)

	expectedResult := fmt.Sprintf(" FAIL  %s\t./%s\n\t- %s\n\n\t\t- asserts[0] `equal` fail\n\t\t\t%s\n\t\t\t%s\n\n",
		given.DisplayName, given.FilePath, given.TestsResult[0].DisplayName, given.TestsResult[0].AssertsResult[0].FailInfo[0], given.TestsResult[0].AssertsResult[0].FailInfo[1])
	given.Print(testPrinter, 0)

	a := assert.New(t)
	a.Equal(expectedResult, buffer.String())
}

func TestTestSuiteResultPrint(t *testing.T) {
	test := TestSuiteResult{
		DisplayName: "this-test-suite",
		TestsResult: []*TestJobResult{
			{
				DisplayName: "first-test",
			},
			{
				DisplayName: "second-skip-test",
			},
			{},
			nil,
		},
	}
	buffer := new(bytes.Buffer)
	test.Print(printer.NewPrinter(buffer, nil), 0)
	for _, result := range test.TestsResult {
		if result == nil {
			continue
		}
		assert.Contains(t, buffer.String(), fmt.Sprintf("- %s", result.DisplayName))
	}
}

// calculate test suite duration
func TestCalculateTestSuiteDuration_NoTests(t *testing.T) {
	cases := []struct {
		name     string
		input    []*TestJobResult
		expected time.Duration
	}{
		{
			name:     "no tests",
			input:    []*TestJobResult{},
			expected: time.Duration(0),
		},
		{
			name: "single test",
			input: []*TestJobResult{
				{Duration: 2 * time.Millisecond},
			},
			expected: 2 * time.Millisecond,
		},
		{
			name: "multiple tests",
			input: []*TestJobResult{
				{Duration: 1 * time.Millisecond},
				{Duration: 2 * time.Millisecond},
				{Duration: 3 * time.Millisecond},
			},
			expected: 6 * time.Millisecond,
		},
		{
			name: "mixed durations",
			input: []*TestJobResult{
				{Duration: 5 * time.Millisecond},
				{Duration: 150 * time.Millisecond},
				{Duration: 2 * time.Microsecond},
			},
			expected: 155*time.Millisecond + 2*time.Microsecond,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tsr := TestSuiteResult{
				TestsResult: tt.input,
			}
			result := tsr.CalculateTestSuiteDuration()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// snapshots counting
func TestCountSnapshot_AllCountsZero(t *testing.T) {
	cache := snapshot.Cache{}
	tsr := TestSuiteResult{
		SnapshotCounting: struct {
			Total    uint
			Failed   uint
			Created  uint
			Vanished uint
		}{
			Total:    0,
			Failed:   0,
			Created:  0,
			Vanished: 0,
		},
	}
	tsr.CountSnapshot(&cache)

	assert.Equal(t, uint(0), tsr.SnapshotCounting.Created)
	assert.Equal(t, uint(0), tsr.SnapshotCounting.Failed)
	assert.Equal(t, uint(0), tsr.SnapshotCounting.Total)
	assert.Equal(t, uint(0), tsr.SnapshotCounting.Vanished)
}

func TestTestSuiteResultPrint_SuiteSkipped(t *testing.T) {
	test := TestSuiteResult{
		DisplayName: "this-test-suite",
		Skipped:     true,
	}
	buffer := new(bytes.Buffer)
	test.Print(printer.NewPrinter(buffer, nil), 0)

	assert.Contains(t, buffer.String(), "SKIP  this-test-suite")
}

func TestTestSuiteResultPrint_TestSkipped(t *testing.T) {
	test := TestSuiteResult{
		DisplayName: "this-test-suite",
		TestsResult: []*TestJobResult{
			{
				DisplayName: "first-test",
			},
			{
				DisplayName: "second-skip-test",
				Skipped:     true,
			},
			{
				DisplayName: "third-test",
			},
		},
	}
	buffer := new(bytes.Buffer)
	test.Print(printer.NewPrinter(buffer, nil), 0)
	assert.NotContains(t, buffer.String(), "SKIP  this-test-suite")
	assert.Contains(t, buffer.String(), "- SKIPPED 'second-skip-test'")
}
