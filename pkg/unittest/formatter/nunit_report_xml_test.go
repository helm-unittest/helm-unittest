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

var tmpNunitTestDir, _ = ioutil.TempDir("", testSuiteTests)

func createNUnitTestCase(name, description, failureContent string, executed bool) NUnitTestCase {
	testCase := NUnitTestCase{
		Name:        name,
		Description: description,
		Success:     "true",
		Asserts:     "0",
		Result:      "Success",
	}

	if len(failureContent) > 0 {
		testCase.Failure = &NUnitFailure{
			Message:    "Failed",
			StackTrace: failureContent,
		}
		testCase.Success = "false"
		testCase.Result = "Failed"
	}

	if executed {
		testCase.Executed = "true"
	} else {
		testCase.Executed = "false"
	}

	return testCase
}

func validateNUnitTestSuite(assert *assert.Assertions, expected, actual []NUnitTestSuite) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Description, actual[i].Description)
			assert.Equal(expected[i].Success, actual[i].Success)
			assert.Equal(expected[i].Executed, actual[i].Executed)
			assert.Equal(expected[i].Result, actual[i].Result)

			// Validate the testcases
			validatNUnitTestCase(assert, expected[i].TestCases, actual[i].TestCases)

			// Recursive validation loop.
			validateNUnitTestSuite(assert, expected[i].TestSuites, actual[i].TestSuites)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func validatNUnitTestCase(assert *assert.Assertions, expected, actual []NUnitTestCase) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Description, actual[i].Description)
			assert.Equal(expected[i].Success, actual[i].Success)
			assert.Equal(expected[i].Asserts, actual[i].Asserts)
			assert.Equal(expected[i].Executed, actual[i].Executed)
			assert.Equal(expected[i].Result, actual[i].Result)

			if expected[i].Failure != nil || actual[i].Failure != nil {
				assert.Equal(expected[i].Failure.Message, actual[i].Failure.Message)
				assert.Equal(expected[i].Failure.StackTrace, actual[i].Failure.StackTrace)
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

func TestWriteTestOutputAsNUnitMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "NUnit_Test_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseDisplayName := "TestCaseSucces"

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "true",
				Executed:    "true",
				Result:      "Success",
				TestCases: []NUnitTestCase{
					createNUnitTestCase(
						testCaseDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseDisplayName),
						"",
						true,
					),
				},
			},
		},
		Name:   "helm-unittest",
		Total:  1,
		Errors: 0,
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

	sut := NewNUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)
}

func TestWriteTestOutputAsNUnitWithFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "NUnit_Test_Failure_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "false",
				Executed:    "true",
				Result:      "Failed",
				TestCases: []NUnitTestCase{
					createNUnitTestCase(
						testCaseSuccessDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseSuccessDisplayName),
						"",
						true,
					),
					createNUnitTestCase(
						testCaseFailureDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseFailureDisplayName),
						failureContent,
						true,
					),
				},
			},
		},
		Name:     "helm-unittest",
		Total:    2,
		Errors:   0,
		Failures: 1,
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

	sut := NewNUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)
}

func TestWriteTestOutputAsNUnitWithFailuresAndErrors(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "NUnit_Test_Failure_And_Errors_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	testCaseErrorDisplayName := "TestCaseError"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)
	errorMessage := "Throw an error."
	failureErrorContent := fmt.Sprintf("%s\n%s", errorMessage, failureContent)

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "false",
				Executed:    "true",
				Result:      "Failed",
				TestCases: []NUnitTestCase{
					createNUnitTestCase(
						testCaseSuccessDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseSuccessDisplayName),
						"",
						true,
					),
					createNUnitTestCase(
						testCaseFailureDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseFailureDisplayName),
						failureContent,
						true,
					),
					createNUnitTestCase(
						testCaseErrorDisplayName,
						fmt.Sprintf("%s.%s", testSuiteDisplayName, testCaseErrorDisplayName),
						failureErrorContent,
						false,
					),
				},
			},
		},
		Name:     "helm-unittest",
		Total:    3,
		Errors:   1,
		Failures: 1,
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

	sut := NewNUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)
}

func TestWriteTestOutputAsNUnitWithErrors(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "NUnit_Test_Errors_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	errorMessage := "Throw an error."

	expected := NUnitTestResults{
		Environment: NUnitEnvironment{},
		CultureInfo: NUnitCultureInfo{},
		TestSuite: []NUnitTestSuite{
			{
				Type:        TestFixture,
				Name:        testSuiteDisplayName,
				Description: outputFile,
				Success:     "false",
				Executed:    "false",
				Result:      "Failed",
				Failure: &NUnitFailure{
					Message:    "Error",
					StackTrace: errorMessage,
				},
			},
		},
		Name:     "helm-unittest",
		Total:    1,
		Errors:   1,
		Failures: 0,
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      false,
			ExecError:   fmt.Errorf("%s", errorMessage),
		},
	}

	sut := NewNUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual NUnitTestResults
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Total, actual.Total)
	assert.Equal(expected.Errors, actual.Errors)
	assert.Equal(expected.Failures, actual.Failures)
	validateNUnitTestSuite(assert, expected.TestSuite, actual.TestSuite)
}
