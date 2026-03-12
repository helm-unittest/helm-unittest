package formatter_test

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
)

func createSonarTestCase(name, duration, message, reason string, isSkipped, isError, isFailed bool) SonarTestCase {
	testCase := SonarTestCase{
		Name:     name,
		Duration: duration,
	}

	if isSkipped {
		testCase.Skipped = &SonarSkipped{
			Message: message,
			Reason:  reason,
		}
	}

	if isError {
		testCase.Error = &SonarError{
			Message:    message,
			Stacktrace: reason,
		}
	}

	if isFailed {
		testCase.Failure = &SonarFailure{
			Message:    message,
			Stacktrace: reason,
		}
	}

	return testCase
}

func validateSonarFiles(assert *assert.Assertions, expected, actual []SonarFile) {
	assert.Equal(len(expected), len(actual))

	for i := range actual {
		assert.Equal(expected[i].Path, actual[i].Path)
		validateSonarTestCases(assert, expected[i].TestCases, actual[i].TestCases)
	}
}

func validateSonarTestCases(assert *assert.Assertions, expected, actual []SonarTestCase) {
	assert.Equal(len(expected), len(actual))

	for i := range actual {
		assert.Equal(expected[i].Name, actual[i].Name)
		assert.Equal(expected[i].Duration, actual[i].Duration)

		if expected[i].Skipped != nil {
			assert.NotNil(actual[i].Skipped)
			assert.Equal(expected[i].Skipped.Message, actual[i].Skipped.Message)
			assert.Equal(expected[i].Skipped.Reason, actual[i].Skipped.Reason)
		}

		if expected[i].Error != nil {
			assert.NotNil(actual[i].Error)
			assert.Equal(expected[i].Error.Message, actual[i].Error.Message)
			assert.Equal(expected[i].Error.Stacktrace, actual[i].Error.Stacktrace)
		}

		if expected[i].Failure != nil {
			assert.NotNil(actual[i].Failure)
			assert.Equal(expected[i].Failure.Message, actual[i].Failure.Message)
			assert.Equal(expected[i].Failure.Stacktrace, actual[i].Failure.Stacktrace)
		}
	}
}

func TestWriteTestOutputAsSonarNoTests(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpNunitTestDir, "Sonar_NoTests_Output.xml")

	expected := SonarTestExecutions{
		Version: 1,
	}

	var given []*results.TestSuiteResult

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	assert.Nil(err)

	assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(assert, expected.Files, actual.Files)
}

func TestWriteTestOutputAsSonarMinimalSuccess(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpNunitTestDir, "Sonar_MinimalSuccess_Output.xml")
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
						"",
						"",
						false,
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
				createTestJobResult(testCaseDisplayName, "", true, false, nil),
			},
		},
	}
	given[0].TestsResult[0].Duration, _ = time.ParseDuration("12s")

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	assert.Nil(err)

	assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(assert, expected.Files, actual.Files)
}

func TestWriteTestOutputAsSonarWithErrorsAndFailures(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpNunitTestDir, "Sonar_ErrorsAndFailures_Output.xml")
	testSuiteDisplayName := "TestingSuiteFailuresAndErrors"
	testCaseDisplayNameError := "TestCaseError"
	testCaseDisplayNameFailure := "TestCaseFailure"
	assertionFailure := "AssertionFailure"
	assertionType := "equal"
	failureMessage := fmt.Sprintf("\t\t - asserts[0] `%s` fail \n\t\t\t %s \n", assertionType, assertionFailure)

	expected := SonarTestExecutions{
		Files: []SonarFile{
			{
				Path: outputFile,
				TestCases: []SonarTestCase{
					createSonarTestCase(
						testCaseDisplayNameError,
						"123",
						"Error",
						"TheError",
						false,
						true,
						false,
					),
					createSonarTestCase(
						testCaseDisplayNameFailure,
						"456",
						"Failed",
						failureMessage,
						false,
						false,
						true,
					),
				},
			},
		},
		Version: 1,
	}

	assertionResults := []*results.AssertionResult{
		createAssertionResult(0, false, false, false, assertionType, assertionFailure, "", ""),
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      true,
			TestsResult: []*results.TestJobResult{
				createTestJobResult(testCaseDisplayNameError, "TheError", false, false, nil),
				createTestJobResult(testCaseDisplayNameFailure, "", false, false, assertionResults),
			},
		},
	}
	given[0].TestsResult[0].Duration, _ = time.ParseDuration("123ms")
	given[0].TestsResult[1].Duration, _ = time.ParseDuration("456ms")

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	assert.Nil(err)

	assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(assert, expected.Files, actual.Files)
}

func TestWriteTestOutputAsSonarWithSkipped(t *testing.T) {
	assert := assert.New(t)
	outputFile := filepath.Join(tmpNunitTestDir, "Sonar_Skipped_Output.xml")
	testSuiteDisplayName := "TestingSuiteSkipped"
	testCaseDisplayNameSkipped := "TestCaseSkipped"
	skipReason := "Version mismatch"
	skippedContent := fmt.Sprintf("SKIPPED '%s'\n\t\t\t %s \n", testCaseDisplayNameSkipped, skipReason)

	expected := SonarTestExecutions{
		Files: []SonarFile{
			{
				Path: outputFile,
				TestCases: []SonarTestCase{
					createSonarTestCase(
						testCaseDisplayNameSkipped,
						"1",
						"Skipped",
						skippedContent,
						true,
						false,
						false,
					),
				},
			},
		},
		Version: 1,
	}

	assertionSkippedResults := []*results.AssertionResult{
		createAssertionResult(0, false, true, false, "", "", skipReason, ""),
	}

	given := []*results.TestSuiteResult{
		{
			DisplayName: testSuiteDisplayName,
			FilePath:    outputFile,
			Passed:      true,
			TestsResult: []*results.TestJobResult{
				createTestJobResult(testCaseDisplayNameSkipped, "", false, true, assertionSkippedResults),
			},
		},
	}
	given[0].TestsResult[0].Duration, _ = time.ParseDuration("1ms")

	sut := NewSonarReportXML()
	byteValue := loadFormatterTestcase(assert, outputFile, given, sut)

	var actual SonarTestExecutions
	err := xml.Unmarshal(byteValue, &actual)
	assert.Nil(err)

	assert.Equal(expected.Version, actual.Version)
	validateSonarFiles(assert, expected.Files, actual.Files)
}
