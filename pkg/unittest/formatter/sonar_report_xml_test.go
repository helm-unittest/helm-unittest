package formatter_test

import (
	"encoding/xml"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

func createSonarTestCase(name string, duration string, isError bool, isFailed bool) SonarTestCase {
	testCase := SonarTestCase{
		Name:     name,
		Duration: duration,
	}

	if isError {
		testCase.Error = &SonarError{}
	}

	if isFailed {
		testCase.Failure = &SonarFailure{}
	}

	return testCase
}

func validateSonarFiles(assert *assert.Assertions, expected, actual []SonarFile) {
	assert.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		assert.Equal(expected[i].Path, actual[i].Path)
		validateSonarTestCases(assert, expected[i].TestCases, actual[i].TestCases)
	}
}

func validateSonarTestCases(assert *assert.Assertions, expected, actual []SonarTestCase) {
	assert.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		assert.Equal(expected[i].Name, actual[i].Name)
		assert.Equal(expected[i].Duration, actual[i].Duration)

	}
}

func TestWriteTestOutputAsSonarNoTests(t *testing.T) {
	_assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "Sonar_Test_Output.xml")

	expected := SonarTestExecutions{
		Version: 1,
	}

	var given []*results.TestSuiteResult

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(_assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	if err != nil {
		_assert.Fail(err.Error())
		return
	}

	_assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(_assert, expected.Files, actual.Files)
}

func TestWriteTestOutputAsSonarMinimalSuccess(t *testing.T) {
	_assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "Sonar_Test_Output.xml")
	testSuiteDisplayName := "TestingSuiteSuccess"
	testCaseDisplayName := "TestCaseSuccessSuccess"

	expected := SonarTestExecutions{
		Files: []SonarFile{
			{
				Path: outputFile,
				TestCases: []SonarTestCase{
					createSonarTestCase(
						testCaseDisplayName,
						"12000",
						false,
						false,
					),
				},
			},
		},
		Version: 1,
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
	given[0].TestsResult[0].Duration, _ = time.ParseDuration("12s")

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(_assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	if err != nil {
		_assert.Fail(err.Error())
	}

	_assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(_assert, expected.Files, actual.Files)
}

func TestWriteTestOutputAsSonarWithErrorsAndFailures(t *testing.T) {
	_assert := assert.New(t)
	outputFile := path.Join(tmpNunitTestDir, "Sonar_Test_Output.xml")
	testSuiteDisplayName := "TestingSuiteFailuresAndErrors"
	testCaseDisplayNameError := "TestCaseError"
	testCaseDisplayNameFailure := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"

	expected := SonarTestExecutions{
		Files: []SonarFile{
			{
				Path: outputFile,
				TestCases: []SonarTestCase{
					createSonarTestCase(
						testCaseDisplayNameError,
						"123",
						true,
						false,
					),
					createSonarTestCase(
						testCaseDisplayNameFailure,
						"456",
						false,
						true,
					),
				},
			},
		},
		Version: 1,
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
				createTestJobResult(testCaseDisplayNameError, "TheError", false, nil),
				createTestJobResult(testCaseDisplayNameFailure, "", false, assertionResults),
			},
		},
	}
	given[0].TestsResult[0].Duration, _ = time.ParseDuration("123ms")
	given[0].TestsResult[1].Duration, _ = time.ParseDuration("456ms")

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(_assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	if err != nil {
		_assert.Fail(err.Error())
	}

	_assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(_assert, expected.Files, actual.Files)
}
