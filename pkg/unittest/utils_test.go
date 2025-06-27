package unittest_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
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
	passed := runner.RunV3([]string{"testdata/chart-failed-template"})
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

func TestV3RunnerWith_Fixture_Chart_DocumentSelector(t *testing.T) {
	cases := []struct {
		chart      string
		testFlavor string
		expected   []string
	}{
		{
			chart:      "testdata/chart-document-selector",
			testFlavor: "case1-error",
			expected: []string{
				"### Error:  empty 'documentSelector.path' not supported",
			},
		},
		{
			chart:      "testdata/chart-document-selector",
			testFlavor: "case2-error",
			expected: []string{
				"### Error:  empty 'documentSelector.value' not supported",
			},
		},
		{
			chart:      "testdata/chart-document-selector",
			testFlavor: "case3-error",
			expected: []string{
				"Template:\tdocument-selector/templates/cfg01.yaml",
				"Path:\tkind expected to exist",
			},
		},
		{
			chart:      "testdata/chart-document-selector",
			testFlavor: "case4-error",
			expected: []string{
				"Path:\tkind expected to exists",
			},
		},
		{
			chart:      "testdata/chart-document-selector",
			testFlavor: "case1-ok",
			expected: []string{
				"Test Suites: 2 passed, 2 total",
			},
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("chart %s with %s", tt.chart, tt.testFlavor), func(t *testing.T) {
			buffer := new(bytes.Buffer)
			runner := TestRunner{
				Printer:   printer.NewPrinter(buffer, nil),
				TestFiles: []string{fmt.Sprintf("tests/%s_test.yaml", tt.testFlavor)},
			}
			_ = runner.RunV3([]string{tt.chart})
			for _, e := range tt.expected {
				assert.Contains(t, buffer.String(), e)
			}
		})
	}
}

func TestV3RunnerWith_Fixture_Chart_SkipTest(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    false,
	}
	_ = runner.RunV3([]string{"testdata/chart-skip-test"})
	assert.Contains(t, buffer.String(), "- SKIPPED 'should skip first test'")
	assert.Contains(t, buffer.String(), "- SKIPPED 'should third test'")
	assert.Contains(t, buffer.String(), "Test Suites: 2 passed, 1 skipped, 3 total")
	assert.Contains(t, buffer.String(), "Tests:       2 passed, 4 skipped, 6 total")
}

func TestV3RunnerWith_Fixture_Chart_PostRenderer(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    false,
	}
	_ = runner.RunV3([]string{"testdata/chart-post-renderer"})
	assert.Contains(t, buffer.String(), "Test Suites: 2 passed, 2 total")
	assert.Contains(t, buffer.String(), "Tests:       4 passed, 4 total")
}

func TestV3RunnerWith_Fixture_Chart_WithSubchartV1(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    true,
	}
	_ = runner.RunV3([]string{"testdata/chart-subchart-v1"})

	assert.Contains(t, buffer.String(), "Test Suites: 3 passed, 3 total")
	assert.Contains(t, buffer.String(), "Tests:       9 passed, 9 total")
}

func TestSplitBefore(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator string
		expected  []string
	}{
		// This suite was constructed against a test suite for SplitAfter (to check for parity).
		{
			name:      "Simple Case",
			input:     "apple-banana-cherry",
			separator: "-",
			expected:  []string{"apple", "-banana", "-cherry"}, // Separator at the beginning
		},
		{
			name:      "No Separator",
			input:     "apple",
			separator: "-",
			expected:  []string{"apple"},
		},
		{
			name:      "Separator at Beginning",
			input:     "-apple-banana",
			separator: "-",
			expected:  []string{"", "-apple", "-banana"}, // Separator at the beginning
		},
		{
			name:      "Separator at End",
			input:     "apple-banana-",
			separator: "-",
			expected:  []string{"apple", "-banana", "-"}, // Separator at the beginning
		},
		{
			name:      "Consecutive Separators",
			input:     "apple--banana",
			separator: "-",
			expected:  []string{"apple", "-", "-banana"}, // Separator at the beginning
		},
		{
			name:      "Empty String",
			input:     "",
			separator: "-",
			expected:  []string{""},
		},
		{
			name:      "Special Characters",
			input:     "one.two.three",
			separator: ".",
			expected:  []string{"one", ".two", ".three"},
		},
		{
			name:      "Long Separator",
			input:     "prefix-one-suffix-two",
			separator: "-one-",
			expected:  []string{"prefix", "-one-suffix-two"},
		},
		{
			name:      "Separator and other special characters",
			input:     "prefix.one-suffix.two",
			separator: ".",
			expected:  []string{"prefix", ".one-suffix", ".two"},
		},
		{
			name:      "Complex Case",
			input:     "part1-part2-SEP-part3-part4-SEP-part5",
			separator: "SEP",
			expected:  []string{"part1-part2-", "SEP-part3-part4-", "SEP-part5"},
		},
		{
			name:      "Complex Case with SEP at beginning and end",
			input:     "SEP-part1-part2-SEP-part3-part4-SEP",
			separator: "SEP",
			expected:  []string{"", "SEP-part1-part2-", "SEP-part3-part4-", "SEP"},
		},
		{
			name:      "Another Complex Case",
			input:     "before#### file: test1\nmanifest1\n#### file: test2\nmanifest2",
			separator: "#### file:",
			expected:  []string{"before", "#### file: test1\nmanifest1\n", "#### file: test2\nmanifest2"},
		},
		{
			name:      "Empty Input with Separator at Beginning",
			input:     "#### file:",
			separator: "#### file:",
			expected:  []string{"", "#### file:"},
		},
		{
			name:      "Input with only Separator",
			input:     "#### file:",
			separator: "#### file:",
			expected:  []string{"", "#### file:"},
		},
		{
			name:      "Input with only Separator repeated",
			input:     "#### file:#### file:",
			separator: "#### file:",
			expected:  []string{"", "#### file:", "#### file:"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := common.SplitBefore(test.input, test.separator)
			assert.Equal(t, test.expected, actual, fmt.Sprintf("Test Case: %s", test.name))
		})
	}
}

func TestV3RunnerWith_Fixture_Chart_SkipEmptyTemplateWhenEmpty(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    true,
	}
	_ = runner.RunV3([]string{"testdata/chart-skipemptytemplate-no-match"})

	assert.Contains(t, buffer.String(), "Test Suites: 1 passed, 1 total")
	assert.Contains(t, buffer.String(), "Tests:       2 passed, 2 total")
}

func TestV3RunnerWith_Fixture_Chart_SpecialCharacters(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    true,
	}
	_ = runner.RunV3([]string{"testdata/chart-special-characters"})

	assert.Contains(t, buffer.String(), "Test Suites: 1 passed, 1 total")
	assert.Contains(t, buffer.String(), "Tests:       1 passed, 1 total")
}

func TestV3RunnerWith_Fixture_Chart_WithSnapshot_Success(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/success/*_test.yaml"},
		Strict:    true,
	}
	_ = runner.RunV3([]string{"testdata/chart-snapshot"})

	assert.Contains(t, buffer.String(), "Tests:       2 passed, 2 total")
	assert.Contains(t, buffer.String(), "Snapshot:    5 passed, 5 total")
}

func TestV3RunnerWith_Fixture_Chart_WithSnapshot_Failed(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/failed/*_test.yaml"},
		Strict:    true,
	}
	_ = runner.RunV3([]string{"testdata/chart-snapshot"})

	failMsg := `
Template:	snapshot/templates/network.yaml
DocumentIndex:	0
ValuesIndex:	0
Expected pattern '.*not-in-snapshot.*' not found in snapshot:
	app: test-cluster
	app.kubernetes.io/version: null
`

	assert.Contains(t, strings.Join(strings.Fields(buffer.String()), ""), strings.Join(strings.Fields(failMsg), ""))

	assert.Contains(t, buffer.String(), "Tests:       1 failed, 0 passed, 1 total")
	assert.Contains(t, buffer.String(), "Snapshot:    1 passed, 1 total")
}
