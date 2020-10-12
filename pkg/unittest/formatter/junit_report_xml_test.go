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

var tmpJUnitTestDir, _ = ioutil.TempDir("", testSuiteTests)

func createJUnitTestCase(classname, name, failureContent string) JUnitTestCase {
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
	outputFile := path.Join(tmpJUnitTestDir, "JUnit_Test_Output.xml")
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
					createJUnitTestCase(testSuiteDisplayName, testCaseDisplayName, ""),
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
	outputFile := path.Join(tmpJUnitTestDir, "JUnit_Test_Failure_Output.xml")
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
					createJUnitTestCase(testSuiteDisplayName, testCaseSuccessDisplayName, ""),
					createJUnitTestCase(testSuiteDisplayName, testCaseFailureDisplayName, failureContent),
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
			Passed:      true,
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
