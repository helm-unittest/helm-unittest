package results_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/lrills/helm-unittest/internal/printer"
	. "github.com/lrills/helm-unittest/pkg/unittest/results"
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
