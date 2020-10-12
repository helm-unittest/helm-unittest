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
						FailInfo:   nil,
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
	given := createTestSuiteResult("A Test Suite", "A Test Case", "filePath", "", nil, nil, true, false)

	expectedResult := fmt.Sprintf(" PASS  %s\t./%s\n", given.DisplayName, given.FilePath)
	given.Print(testPrinter, 0)

	a := assert.New(t)
	a.Equal(expectedResult, buffer.String())
}
