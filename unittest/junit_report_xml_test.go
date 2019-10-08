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

var tmpTestDir, _ = ioutil.TempDir("", "_suite_tests")

func TestWriteTestOutputMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpTestDir, "Test_Output.xml")
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
	sut.WriteTestOutput(given, false, writer)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Suites[0].Tests, actual.Suites[0].Tests)
	assert.Equal(expected.Suites[0].Failures, actual.Suites[0].Failures)
	assert.Equal(expected.Suites[0].Name, actual.Suites[0].Name)
	assert.Equal(expected.Suites[0].Properties[0].Name, actual.Suites[0].Properties[0].Name)
	assert.Equal(expected.Suites[0].Properties[0].Value, actual.Suites[0].Properties[0].Value)
	assert.Equal(expected.Suites[0].TestCases[0].Classname, actual.Suites[0].TestCases[0].Classname)
	assert.Equal(expected.Suites[0].TestCases[0].Name, actual.Suites[0].TestCases[0].Name)

	testResult.Close()
	os.Remove(outputFile)
}

func TestWriteTestOutputWithFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := path.Join(tmpTestDir, "Test_Failure_Output.xml")
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
	sut.WriteTestOutput(given, false, writer)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	var actual JUnitTestSuites
	xml.Unmarshal(bytevalue, &actual)

	assert.Equal(expected.Suites[0].Tests, actual.Suites[0].Tests)
	assert.Equal(expected.Suites[0].Failures, actual.Suites[0].Failures)
	assert.Equal(expected.Suites[0].Name, actual.Suites[0].Name)
	assert.Equal(expected.Suites[0].Properties[0].Name, actual.Suites[0].Properties[0].Name)
	assert.Equal(expected.Suites[0].Properties[0].Value, actual.Suites[0].Properties[0].Value)
	assert.Equal(expected.Suites[0].TestCases[0].Classname, actual.Suites[0].TestCases[0].Classname)
	assert.Equal(expected.Suites[0].TestCases[0].Name, actual.Suites[0].TestCases[0].Name)
	assert.Equal(expected.Suites[0].TestCases[1].Classname, actual.Suites[0].TestCases[1].Classname)
	assert.Equal(expected.Suites[0].TestCases[1].Name, actual.Suites[0].TestCases[1].Name)
	assert.Equal(expected.Suites[0].TestCases[1].Failure.Message, actual.Suites[0].TestCases[1].Failure.Message)
	assert.Equal(expected.Suites[0].TestCases[1].Failure.Type, actual.Suites[0].TestCases[1].Failure.Type)
	assert.Equal(expected.Suites[0].TestCases[1].Failure.Contents, actual.Suites[0].TestCases[1].Failure.Contents)

	testResult.Close()
	os.Remove(outputFile)
}
