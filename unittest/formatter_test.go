package unittest_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

func TestNewFormatterWithEmptyOutputFile(t *testing.T) {
	given := ""
	sut := NewFormatter(given, given)
	assert.Nil(t, sut)
}

func TestNewFormatterWithOutputFileAndEmptyOutputType(t *testing.T) {
	outputFile := "../__fixtures__/test-output/test_output.xml"
	given := ""
	sut := NewFormatter(outputFile, given)
	assert.Nil(t, sut)
}

func TestNewFormatterWithOutputFileAndOutputTypeJUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "Junit"
	given := "../__fixtures__/test-output/test_output.xml"
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}

func TestNewFormatterWithOutputFileAndOutputTypeNUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "NUnit"
	given := "../__fixtures__/test-output/test_output.xml"
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}

func TestNewFormatterWithOutputFileAndOutputTypeXUnit(t *testing.T) {
	assert := assert.New(t)
	outputType := "XUnit"
	given := "../__fixtures__/test-output/test_output.xml"
	givenDirectory := filepath.Dir(given)
	defer os.Remove(givenDirectory)
	sut := NewFormatter(given, outputType)
	assert.NotNil(sut)
	assert.DirExists(givenDirectory)
}
