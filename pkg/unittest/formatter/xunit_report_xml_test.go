package formatter_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	. "github.com/lrills/helm-unittest/pkg/unittest/formatter"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
)

var tmpXunitTestDir, _ = ioutil.TempDir("", testSuiteTests)

func createXUnitTestCase(name, description, failureContent string, isError bool) XUnitTestCase {
	testCase := XUnitTestCase{
		Name:   name,
		Type:   description,
		Method: XUnitValidationMethod,
		Result: "Pass",
	}

	if len(failureContent) > 0 {
		testCase.Failure = &XUnitFailure{
			ExceptionType: XUnitValidationMethod,
			Message: &XUnitFailureMessage{
				Data: "Failed",
			},
			StackTrace: &XUnitFailureStackTrace{
				Data: failureContent,
			},
		}
		testCase.Result = "Fail"
	}

	if isError {
		testCase.Failure.ExceptionType = fmt.Sprintf("%s-%s", XUnitValidationMethod, "Error")
	}

	return testCase
}

func assertXUnitTestAssemblies(assert *assert.Assertions, expected, actual []XUnitAssembly) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].ConfigFile, actual[i].ConfigFile)
			assert.Equal(expected[i].TotalTests, actual[i].TotalTests)
			assert.Equal(expected[i].PassedTests, actual[i].PassedTests)
			assert.Equal(expected[i].FailedTests, actual[i].FailedTests)
			assert.Equal(expected[i].SkippedTests, actual[i].SkippedTests)
			assert.Equal(expected[i].ErrorsTests, actual[i].ErrorsTests)

			// Validate the tesruns
			assertXUnitTestRun(assert, expected[i].TestRuns, actual[i].TestRuns)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func assertXUnitTestRun(assert *assert.Assertions, expected, actual []XUnitTestRun) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].TotalTests, actual[i].TotalTests)
			assert.Equal(expected[i].PassedTests, actual[i].PassedTests)
			assert.Equal(expected[i].FailedTests, actual[i].FailedTests)
			assert.Equal(expected[i].SkippedTests, actual[i].SkippedTests)

			// Validate the testcases
			assertXUnitTestCase(assert, expected[i].TestCases, actual[i].TestCases)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func assertXUnitTestCase(assert *assert.Assertions, expected, actual []XUnitTestCase) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Type, actual[i].Type)
			assert.Equal(expected[i].Method, actual[i].Method)
			assert.Equal(expected[i].Result, actual[i].Result)

			if expected[i].Failure != nil || actual[i].Failure != nil {
				assert.Equal(expected[i].Failure.ExceptionType, actual[i].Failure.ExceptionType)
				assert.Equal(expected[i].Failure.Message.Data, actual[i].Failure.Message.Data)
				assert.Equal(expected[i].Failure.StackTrace.Data, actual[i].Failure.StackTrace.Data)
			} else {
				// Verify if both are nil, otherwise it's still a failure.
				assert.True(expected[i].Failure == nil && actual[i].Failure == nil)
			}
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func TestWriteTestOutputAsXUnitMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpXunitTestDir, "XUnit_Test_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseDisplayName := "TestCaseSucces"
	totalTests := 1
	totalPassed := 1
	totalFailed := 0
	totalErrors := 0
	totalSkipped := 0

	expected := XUnitAssemblies{
		Assembly: []XUnitAssembly{
			{
				Name:         outputFile,
				ConfigFile:   outputFile,
				TotalTests:   totalTests,
				PassedTests:  totalPassed,
				FailedTests:  totalFailed,
				SkippedTests: totalSkipped,
				ErrorsTests:  totalErrors,
				TestRuns: []XUnitTestRun{
					{
						Name:         testSuiteDisplayName,
						TotalTests:   totalTests,
						PassedTests:  totalPassed,
						FailedTests:  totalFailed,
						SkippedTests: totalSkipped,
						TestCases: []XUnitTestCase{
							createXUnitTestCase(testCaseDisplayName, testSuiteDisplayName, "", false),
						},
					},
				},
			},
		},
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      true,
			TestsResult: []*results.TestJobResult{
				createTestJobResult(testCaseDisplayName, "", true, nil),
			},
		},
	}

	sut := NewXUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)
}

func TestWriteTestOutputAsXUnitWithFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpXunitTestDir, "XUnit_Test_Failure_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)
	totalTests := 2
	totalPassed := 1
	totalFailed := 1
	totalErrors := 0
	totalSkipped := 0

	expected := XUnitAssemblies{
		Assembly: []XUnitAssembly{
			{
				Name:         outputFile,
				ConfigFile:   outputFile,
				TotalTests:   totalTests,
				PassedTests:  totalPassed,
				SkippedTests: totalSkipped,
				FailedTests:  totalFailed,
				ErrorsTests:  totalErrors,
				TestRuns: []XUnitTestRun{
					{
						Name:         testSuiteDisplayName,
						TotalTests:   totalTests,
						PassedTests:  totalPassed,
						FailedTests:  totalFailed,
						SkippedTests: totalSkipped,
						TestCases: []XUnitTestCase{
							createXUnitTestCase(testCaseSuccessDisplayName, testSuiteDisplayName, "", false),
							createXUnitTestCase(testCaseFailureDisplayName, testSuiteDisplayName, failureContent, false),
						},
					},
				},
			},
		},
	}

	assertionResults := []*results.AssertionResult{
		createAssertionResult(0, false, false, assertionType, assertionFailure, ""),
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			TestsResult: []*results.TestJobResult{
				createTestJobResult(testCaseSuccessDisplayName, "", true, nil),
				createTestJobResult(testCaseFailureDisplayName, "", false, assertionResults),
			},
		},
	}

	sut := NewXUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)
}

func TestWriteTestOutputAsXUnitWithFailuresAndErrors(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpXunitTestDir, "XUnit_Test_Failure_And_Error_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	testCaseErrorDisplayName := "TestCaseError"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)
	errorMessage := "An Error Occurred."
	failureErrorContent := fmt.Sprintf("%s\n%s", errorMessage, failureContent)
	totalTests := 3
	totalPassed := 1
	totalFailed := 1
	totalErrors := 1
	totalSkipped := 0

	expected := XUnitAssemblies{
		Assembly: []XUnitAssembly{
			{
				Name:         outputFile,
				ConfigFile:   outputFile,
				TotalTests:   totalTests,
				PassedTests:  totalPassed,
				SkippedTests: totalSkipped,
				FailedTests:  totalFailed,
				ErrorsTests:  totalErrors,
				TestRuns: []XUnitTestRun{
					{
						Name:         testSuiteDisplayName,
						TotalTests:   totalTests,
						PassedTests:  totalPassed,
						FailedTests:  totalFailed,
						SkippedTests: totalSkipped,
						TestCases: []XUnitTestCase{
							createXUnitTestCase(testCaseSuccessDisplayName, testSuiteDisplayName, "", false),
							createXUnitTestCase(testCaseFailureDisplayName, testSuiteDisplayName, failureContent, false),
							createXUnitTestCase(testCaseErrorDisplayName, testSuiteDisplayName, failureErrorContent, true),
						},
					},
				},
			},
		},
	}

	assertionResults := []*results.AssertionResult{
		createAssertionResult(0, false, false, assertionType, assertionFailure, ""),
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			TestsResult: []*results.TestJobResult{
				createTestJobResult(testCaseSuccessDisplayName, "", true, nil),
				createTestJobResult(testCaseFailureDisplayName, "", false, assertionResults),
				createTestJobResult(testCaseErrorDisplayName, errorMessage, false, assertionResults),
			},
		},
	}

	sut := NewXUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)
}

func TestWriteTestOutputAsXUnitWithErrors(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpXunitTestDir, "XUnit_Test_Error_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	errorMessage := "An Error Occurred."
	totalTests := 1
	totalPassed := 0
	totalFailed := 0
	totalErrors := 1
	totalSkipped := 0

	expected := XUnitAssemblies{
		Assembly: []XUnitAssembly{
			{
				Name:         outputFile,
				ConfigFile:   outputFile,
				TotalTests:   totalTests,
				PassedTests:  totalPassed,
				SkippedTests: totalSkipped,
				FailedTests:  totalFailed,
				ErrorsTests:  totalErrors,
				Errors: []XUnitError{
					{
						Type: "Error",
						Name: "Error",
						Failure: &XUnitFailure{
							ExceptionType: fmt.Sprintf("%s-%s", XUnitValidationMethod, "Error"),
							Message: &XUnitFailureMessage{
								Data: "Failed",
							},
							StackTrace: &XUnitFailureStackTrace{
								Data: errorMessage,
							},
						},
					},
				},
			},
		},
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			ExecError:   fmt.Errorf("%s", errorMessage),
		},
	}

	sut := NewXUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)
}
