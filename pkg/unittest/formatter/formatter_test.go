package formatter_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/lrills/helm-unittest/pkg/unittest/formatter"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
)

const testSuiteTests string = "_suite_tests"
const testOutputFile string = "../../../test/data/output/test_output.xml"

func createTestJobResult(name, errorMessage string, passed bool, assertionResults []*results.AssertionResult) *results.TestJobResult {
	testJobResult := &results.TestJobResult{
		DisplayName: name,
		Passed:      passed,
	}

	if len(errorMessage) > 0 {
		testJobResult.ExecError = fmt.Errorf("%s", errorMessage)
	}

	if assertionResults != nil {
		testJobResult.AssertsResult = assertionResults
	}

	return testJobResult
}

func createAssertionResult(index int, passed, not bool, assertionType, failInfo, customInfo string) *results.AssertionResult {
	return &results.AssertionResult{
		Index:      index,
		FailInfo:   []string{failInfo},
		Passed:     passed,
		AssertType: assertionType,
		Not:        not,
		CustomInfo: customInfo,
	}
}

func loadFormatterTestcase(assert *assert.Assertions, outputFile string, given []*results.TestSuiteResult, sut Formatter) []byte {

	writer, cerr := os.Create(outputFile)
	assert.Nil(cerr)

	// Test the formatter
	serr := sut.WriteTestOutput(given, false, writer)
	assert.Nil(serr)

	// Don't defer, as we want to close it before stopping the test.
	writer.Close()

	assert.FileExists(outputFile)

	// Unmarshall and validate the output with expected.
	testResult, rerr := os.Open(outputFile)
	assert.Nil(rerr)
	bytevalue, _ := ioutil.ReadAll(testResult)

	testResult.Close()
	os.Remove(outputFile)

	return bytevalue
}

func TestNewFormatterWithEmptyOutputFile(t *testing.T) {
	given := ""
	sut := NewFormatter(given, given)
	assert.Nil(t, sut)
}

func TestNewFormatterWithOutputFileAndEmptyOutputType(t *testing.T) {
	given := ""
	sut := NewFormatter(testOutputFile, given)
	assert.Nil(t, sut)
}

func TestNewFormatterWithOutputFileAndOutputTypeJUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "Junit"
	given := testOutputFile
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}

func TestNewFormatterWithOutputFileAndOutputTypeNUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "NUnit"
	given := testOutputFile
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}

func TestNewFormatterWithOutputFileAndOutputTypeXUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "XUnit"
	given := testOutputFile
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}
