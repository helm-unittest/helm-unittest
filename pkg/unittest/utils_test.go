package unittest_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/internal/printer"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/stretchr/testify/assert"
)

// unmarshalJobTestHelper unmarshall a YAML-encoded string into a TestJob struct.
// It extracts the majorVersion, minorVersion, and apiVersions fields from
// CapabilitiesFields and populates the corresponding fields in Capabilities.
// If apiVersions is nil, it sets APIVersions to nil. If it's a slice,
// it appends string values to APIVersions. Returns an error if unmarshaling
// or processing fails.
func unmarshalJobTestHelper(input string, out *TestJob, t *testing.T) {
	t.Helper()
	err := common.YmlUnmarshal(input, &out)
	assert.NoError(t, err)
	out.SetCapabilities()
}

// writeToFile writes the provided string data to a file with the given filename.
// It returns an error if the file cannot be created or if there is an error during writing.
func writeToFile(data string, filename string) error {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}

	// Create the file with an absolute path
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

// This file contains unit tests for the TestRunner functionality in the unittest package.
// The purpose of these tests is to verify that the TestRunner behaves correctly when running
// tests on Helm charts, especially when handling multiple complex cases in the test files.
// The tests ensure that the TestRunner can correctly process and report the results of the tests.

// How to add more end-2-end tests
// 1. Create a new function called TestV3RunnerWith_Fixture_Chart_<Context>
// 2. Create fixtures in the `testdata` directory. Example `testdata/chart<number>`
// 3. Create test files in the `tests` directory. Example `testdata/chart<number>/<name>_test.yaml`
// 4. Add metadata information to the test files. Example `testdata/chart<number>/Chart.yaml`

func TestV3RunnerWith_Fixture_Chart_ErrorWhenMetaCharacters(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
	}
	passed := runner.RunV3([]string{"testdata/chart01"})
	assert.True(t, passed, buffer.String())
}

func TestV3RunnerWith_Fixture_Chart_FailFast(t *testing.T) {
	cases := []struct {
		chart      string
		failFast   bool
		testFlavor string
		expected   []string
	}{
		{
			chart:      "testdata/chart-fail-fast",
			failFast:   true,
			testFlavor: "case1",
			expected: []string{
				"FAIL  a fail-fast first test",
				"Test Suites: 1 failed, 0 passed, 1 total",
				"Tests:       1 failed, 1 passed, 2 total",
			},
		},
		{
			chart:      "testdata/chart-fail-fast",
			failFast:   false,
			testFlavor: "case1",
			expected: []string{
				"FAIL  a fail-fast first test",
				"PASS  b fail-fast second test",
				"Test Suites: 1 failed, 1 passed, 2 total",
				"Tests:       1 failed, 4 passed, 5 total",
			},
		},
		{
			chart:      "testdata/chart-fail-fast",
			failFast:   false,
			testFlavor: "case2",
			expected: []string{
				"PASS  a fail-fast first test all pass",
				"FAIL  b fail-fast second test",
				"Test Suites: 1 failed, 1 passed, 2 total",
				"Tests:       1 failed, 5 passed, 6 total",
			},
		},
		{
			chart:      "testdata/chart-fail-fast",
			failFast:   true,
			testFlavor: "case2",
			expected: []string{
				"PASS  a fail-fast first test all pass",
				"FAIL  b fail-fast second test",
				"Test Suites: 1 failed, 1 passed, 2 total",
				"Tests:       1 failed, 5 passed, 6 total",
			},
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("chart %s with %s fail fast %v", tt.chart, tt.testFlavor, tt.failFast), func(t *testing.T) {
			buffer := new(bytes.Buffer)
			runner := TestRunner{
				Printer:   printer.NewPrinter(buffer, nil),
				TestFiles: []string{fmt.Sprintf("tests/*-%s_test.yaml", tt.testFlavor)},
				Failfast:  tt.failFast,
			}
			_ = runner.RunV3([]string{"testdata/chart-fail-fast"})
			for _, e := range tt.expected {
				assert.Contains(t, buffer.String(), e)
			}
		})
	}
}

func TestV3RunnerWith_Fixture_Chart_YamlSeparator(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    false,
	}
	_ = runner.RunV3([]string{"testdata/chart-yaml-separator"})
	assert.Contains(t, buffer.String(), "Test Suites: 5 passed, 5 total")
	assert.Contains(t, buffer.String(), "Tests:       6 passed, 6 total")
}
