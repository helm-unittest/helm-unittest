package formatter_test

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
)

var tmpJUnitTestDir, _ = os.MkdirTemp("", testSuiteTests)

func createJUnitTestCase(classname, name, failureContent string, errorContent error) JUnitTestCase {
	testCase := JUnitTestCase{
		Classname: classname,
		Name:      name,
	}

	if len(failureContent) > 0 {
		testCase.Failure = &JUnitFailure{
			Message:  "Failed",
			Type:     "",
			Contents: failureContent,
		}
	}

	if errorContent != nil {
		testCase.Error = &JUnitFailure{
			Message:  "Error",
			Type:     "",
			Contents: errorContent.Error(),
		}
	}

	return testCase
}

func createJUnitProperty(name, value string) JUnitProperty {
	return JUnitProperty{
		Name:  name,
		Value: value,
	}
}

func assertJUnitTestSuite(assert *assert.Assertions, expected, actual []JUnitTestSuite) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Tests, actual[i].Tests)
			assert.Equal(expected[i].Errors, actual[i].Errors)
			assert.Equal(expected[i].Failures, actual[i].Failures)
			assert.Equal(expected[i].Name, actual[i].Name)

			assertJUnitProperty(assert, expected[i].Properties, actual[i].Properties)
			assertJUnitTestCase(assert, expected[i].TestCases, actual[i].TestCases)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func assertJUnitTestCase(assert *assert.Assertions, expected, actual []JUnitTestCase) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Classname, actual[i].Classname)
			assert.Equal(expected[i].Name, actual[i].Name)

			if expected[i].Failure != nil && actual[i].Failure != nil {
				assert.Equal(expected[i].Failure.Message, actual[i].Failure.Message)
				assert.Equal(expected[i].Failure.Type, actual[i].Failure.Type)
				assert.Equal(expected[i].Failure.Contents, actual[i].Failure.Contents)
			} else {
				assert.True(expected[i].Failure == nil && actual[i].Failure == nil)
			}

			if expected[i].Error != nil && actual[i].Error != nil {
				assert.Equal(expected[i].Error.Message, actual[i].Error.Message)
				assert.Equal(expected[i].Error.Type, actual[i].Error.Type)
				assert.Equal(expected[i].Error.Contents, actual[i].Error.Contents)
			} else {
				assert.True(expected[i].Error == nil && actual[i].Error == nil)
			}
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func assertJUnitProperty(assert *assert.Assertions, expected, actual []JUnitProperty) {

	if expected != nil && actual != nil {
		actualLength := len(actual)
		assert.Equal(len(expected), actualLength)

		for i := 0; i < actualLength; i++ {
			assert.Equal(expected[i].Name, actual[i].Name)
			assert.Equal(expected[i].Value, actual[i].Value)
		}
	} else {
		// Verify if both are nil, otherwise it's still a failure.
		assert.True(expected == nil && actual == nil)
	}
}

func TestWriteTestOutputAsJUnitMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpJUnitTestDir, "JUnit_Test_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseDisplayName := "TestCaseSucces"

	expected := JUnitTestSuites{
		Suites: []JUnitTestSuite{
			{
				Tests:    1,
				Failures: 0,
				Name:     testSuiteDisplayName,
				Properties: []JUnitProperty{
					createJUnitProperty("helm-unittest.version", "1.6"),
				},
				TestCases: []JUnitTestCase{
					createJUnitTestCase(testSuiteDisplayName, testCaseDisplayName, "", nil),
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

	sut := NewJUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assertJUnitTestSuite(assert, expected.Suites, actual.Suites)
}

func TestWriteTestOutputAsJUnitWithFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpJUnitTestDir, "JUnit_Test_Failure_Output.xml")
	testSuiteDisplayName := "TestingSuite"
	testCaseSuccessDisplayName := "TestCaseSuccess"
	testCaseFailureDisplayName := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	assertIndex := 0
	failureContent := fmt.Sprintf("\t\t - asserts[%d]%s `%s` fail \n\t\t\t %s \n", assertIndex, "", assertionType, assertionFailure)

	expected := JUnitTestSuites{
		Suites: []JUnitTestSuite{
			{
				Tests:    2,
				Failures: 1,
				Name:     testSuiteDisplayName,
				Properties: []JUnitProperty{
					createJUnitProperty("helm-unittest.version", "1.6"),
				},
				TestCases: []JUnitTestCase{
					createJUnitTestCase(testSuiteDisplayName, testCaseSuccessDisplayName, "", nil),
					createJUnitTestCase(testSuiteDisplayName, testCaseFailureDisplayName, failureContent, nil),
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

	sut := NewJUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assertJUnitTestSuite(assert, expected.Suites, actual.Suites)
}

func TestWriteTestOutputAsJUnitWithFailuresAndErrors(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpJUnitTestDir, "JUnit_Test_FailureError_Output.xml")
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

	expected := JUnitTestSuites{
		Suites: []JUnitTestSuite{
			{
				Tests:    3,
				Errors:   1,
				Failures: 1,
				Name:     testSuiteDisplayName,
				Properties: []JUnitProperty{
					createJUnitProperty("helm-unittest.version", "1.6"),
				},
				TestCases: []JUnitTestCase{
					createJUnitTestCase(testSuiteDisplayName, testCaseSuccessDisplayName, "", nil),
					createJUnitTestCase(testSuiteDisplayName, testCaseFailureDisplayName, failureContent, nil),
					createJUnitTestCase(testSuiteDisplayName, testCaseErrorDisplayName, "", fmt.Errorf("%s", failureErrorContent)),
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
				createTestJobResult(testCaseErrorDisplayName, failureErrorContent, false, nil),
			},
		},
	}

	sut := NewJUnitReportXML()
	bytevalue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assertJUnitTestSuite(assert, expected.Suites, actual.Suites)
}
