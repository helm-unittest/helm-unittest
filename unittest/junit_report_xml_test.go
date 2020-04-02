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

var tmpJUnitTestDir, _ = ioutil.TempDir("", "_suite_tests")

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
					{
						Name:  "helm-unittest.version",
						Value: "1.6",
					},
				},
				TestCases: []JUnitTestCase{
					{
						Classname: testSuiteDisplayName,
						Name:      testCaseDisplayName,
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
	sut := NewJUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assertJUnitTestSuite(assert, expected.Suites, actual.Suites)

	testResult.Close()
	os.Remove(outputFile)
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
					{
						Name:  "helm-unittest.version",
						Value: "1.6",
					},
				},
				TestCases: []JUnitTestCase{
					{
						Classname: testSuiteDisplayName,
						Name:      testCaseSuccessDisplayName,
					},
					{
						Classname: testSuiteDisplayName,
						Name:      testCaseFailureDisplayName,
						Failure: &JUnitFailure{
							Message:  "Failed",
							Type:     "",
							Contents: failureContent,
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
	sut := NewJUnitReportXML()
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assertJUnitTestSuite(assert, expected.Suites, actual.Suites)

	testResult.Close()
	os.Remove(outputFile)
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
