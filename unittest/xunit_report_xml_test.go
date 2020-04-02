package unittest_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

var tmpXunitTestDir, _ = ioutil.TempDir("", "_suite_tests")

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
							{
								Name:    testCaseDisplayName,
								Type:    testSuiteDisplayName,
								Method:  "Helm-Validation",
								Result:  "Pass",
								Failure: nil,
							},
						},
					},
				},
			},
		},
	}

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      true,
			TestsResult: []*TestJobResult{
				{
					DisplayName: testCaseDisplayName,
					Passed:      true,
				},
			},
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewXUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)

	testResult.Close()
	os.Remove(outputFile)
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
							{
								Name:    testCaseSuccessDisplayName,
								Type:    testSuiteDisplayName,
								Method:  "Helm-Validation",
								Result:  "Pass",
								Failure: nil,
							},
							{
								Name:   testCaseFailureDisplayName,
								Type:   testSuiteDisplayName,
								Method: "Helm-Validation",
								Result: "Fail",
								Failure: &XUnitFailure{
									ExceptionType: "Helm-Validation",
									Message: &XUnitFailureMessage{
										Data: "Failed",
									},
									StackTrace: &XUnitFailureStackTrace{
										Data: failureContent,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			TestsResult: []*TestJobResult{
				{
					DisplayName: testCaseSuccessDisplayName,
					Passed:      true,
				},
				{
					DisplayName: testCaseFailureDisplayName,
					Passed:      false,
					AssertsResult: []*AssertionResult{
						{
							Index: 0,
							FailInfo: []string{
								assertionFailure,
							},
							Passed:     false,
							AssertType: assertionType,
							Not:        false,
						},
					},
				},
			},
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewXUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)

	testResult.Close()
	os.Remove(outputFile)
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
							{
								Name:    testCaseSuccessDisplayName,
								Type:    testSuiteDisplayName,
								Method:  "Helm-Validation",
								Result:  "Pass",
								Failure: nil,
							},
							{
								Name:   testCaseFailureDisplayName,
								Type:   testSuiteDisplayName,
								Method: "Helm-Validation",
								Result: "Fail",
								Failure: &XUnitFailure{
									ExceptionType: "Helm-Validation",
									Message: &XUnitFailureMessage{
										Data: "Failed",
									},
									StackTrace: &XUnitFailureStackTrace{
										Data: failureContent,
									},
								},
							},
							{
								Name:   testCaseErrorDisplayName,
								Type:   testSuiteDisplayName,
								Method: "Helm-Validation",
								Result: "Fail",
								Failure: &XUnitFailure{
									ExceptionType: "Helm-Validation",
									Message: &XUnitFailureMessage{
										Data: "Failed",
									},
									StackTrace: &XUnitFailureStackTrace{
										Data: failureErrorContent,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			TestsResult: []*TestJobResult{
				{
					DisplayName: testCaseSuccessDisplayName,
					Passed:      true,
				},
				{
					DisplayName: testCaseFailureDisplayName,
					Passed:      false,
					AssertsResult: []*AssertionResult{
						{
							Index: 0,
							FailInfo: []string{
								assertionFailure,
							},
							Passed:     false,
							AssertType: assertionType,
							Not:        false,
						},
					},
				},
				{
					DisplayName: testCaseErrorDisplayName,
					Passed:      false,
					AssertsResult: []*AssertionResult{
						{
							Index: 0,
							FailInfo: []string{
								assertionFailure,
							},
							Passed:     false,
							AssertType: assertionType,
							Not:        false,
						},
					},
					ExecError: fmt.Errorf("%s", errorMessage),
				},
			},
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewXUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)

	testResult.Close()
	os.Remove(outputFile)
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
							ExceptionType: "Helm-Validation-Error",
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

	given := []*TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			ExecError:   fmt.Errorf("%s", errorMessage),
		},
	}

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	sut := NewXUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual XUnitAssemblies
	xml.Unmarshal(bytevalue, &actual)

	assertXUnitTestAssemblies(assert, expected.Assembly, actual.Assembly)

	testResult.Close()
	os.Remove(outputFile)
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
