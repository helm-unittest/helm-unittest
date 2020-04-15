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
