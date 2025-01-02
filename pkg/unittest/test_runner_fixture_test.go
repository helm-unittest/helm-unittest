package unittest

// This file contains unit tests for the TestRunner functionality in the unittest package.
// The purpose of these tests is to verify that the TestRunner behaves correctly when running
// tests on Helm charts, especially when handling multiple complex cases in the test files.
// The tests ensure that the TestRunner can correctly process and report the results of the tests.

// How to add more end-2-end tests
// 1. Create a new function called TestV3RunnerWith_Fixture_Chart_<Context>
// 2. Create fixtures in the `testdata` directory. Example `testdata/chart<number>`
// 3. Create test files in the `tests` directory. Example `testdata/chart<number>/<name>_test.yaml`
// 4. Add metadata information to the test files. Example `testdata/chart<number>/Chart.yaml`

import (
	"bytes"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/printer"
	"github.com/stretchr/testify/assert"
)

func TestV3RunnerWith_Fixture_Chart_ErrorWhenMetaCharacters(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
	}
	passed := runner.RunV3([]string{"testdata/chart01"})
	assert.True(t, passed, buffer.String())
}
