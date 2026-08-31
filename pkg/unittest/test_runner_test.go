package unittest_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

var sectionBeginPattern = regexp.MustCompile("( PASS | FAIL |\n*###|\n*Charts:|\n*Snapshot Summary:)")
var timePattern = regexp.MustCompile(`(Time:\s+)(?:[\d\.]+)(s|ms|\xB5s)`) // B5 = micron for microseconds

func makeOutputSnapshotable(originalOutput string) []any {
	output := strings.ReplaceAll(originalOutput, "\\", "/")
	timeAgnosticOutput := timePattern.ReplaceAllString(output, "${1}XX.XXXms")

	sectionBeggingLocs := sectionBeginPattern.FindAllStringIndex(timeAgnosticOutput, -1)
	sections := make([]string, len(sectionBeggingLocs))

	suiteBeginIdx := -1
	for sectionIdx := range sections {
		start := sectionBeggingLocs[sectionIdx][0]
		var end int
		if sectionIdx >= len(sections)-1 {
			end = len(timeAgnosticOutput)
		} else {
			end = sectionBeggingLocs[sectionIdx+1][0]
		}

		sectionContent := timeAgnosticOutput[start:end]
		sectionBegin := sectionContent[:6]
		if sectionBegin == " PASS " || sectionBegin == " FAIL " {
			sections[sectionIdx] = strings.TrimRight(sectionContent, "\n")
			if suiteBeginIdx == -1 {
				suiteBeginIdx = sectionIdx
			}
		} else {
			sections[sectionIdx] = sectionContent
			if suiteBeginIdx != -1 {
				sort.Strings(sections[suiteBeginIdx:sectionIdx])
				suiteBeginIdx = -1
			}
		}
	}

	sectionsToRetrun := make([]any, len(sections))
	for idx, section := range sections {
		sectionsToRetrun[idx] = section
	}
	return sectionsToRetrun
}

func TestV4RunnerInvalidChartDirFailfast(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		Failfast:  true,
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testTestFiles})
	assert.False(t, passed, buffer.String())
}

func TestV4RunnerInvalidTestSuiteFailfast(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		Strict:    false,
		Failfast:  true,
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4InvalidBasicChart})
	assert.False(t, passed, buffer.String())
}

func TestV4RunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithPassedTestsDifferentFormatter(t *testing.T) {
	outputFile := "output.txt"
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:    printer.NewPrinter(buffer, nil),
		TestFiles:  []string{testTestFiles},
		OutputFile: outputFile,
		Formatter:  formatter.NewSonarReportXML(),
	}
	passed := runner.RunV4([]string{testV4BasicChart})
	assert.True(t, passed, buffer.String())
	// clean up output file if exists
	if _, err := os.Stat(outputFile); err == nil {
		err = os.Remove(outputFile)
		assert.NoError(t, err)
	} else {
		assert.Fail(t, "Output file not created")
	}
}

func TestV4RunnerOkWithSubSubChartsPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		WithSubChart: true,
		Printer:      printer.NewPrinter(buffer, nil),
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSubSubFolderChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithFailingTemplatePassedTest(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithFailingTemplateChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithOverrideValuesPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testTestFiles},
		ValuesFiles: []string{testValuesFiles},
	}
	passed := runner.RunV4([]string{testV4BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithAbsoluteOverrideValuesPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	fullPath, _ := filepath.Abs(testValuesFiles)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testTestFiles},
		ValuesFiles: []string{fullPath},
	}
	passed := runner.RunV4([]string{testV4BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFailedFiles},
	}
	passed := runner.RunV4([]string{testV4BasicChart})
	assert.False(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithSubSubfolder(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSubFolderChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		WithSubChart: true,
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		WithSubChart: false,
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkGlobalDoubleWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4GlobalDoubleChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithFiles(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithFilesChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithFullsnapshot(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4FullSnapshotChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithRenderedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:        printer.NewPrinter(buffer, nil),
		ChartTestsPath: "tests-chart",
	}
	passed := runner.RunV4([]string{testV4WithHelmTestsChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithDocumentSelector(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithDocumentSelectorChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithDocumentSelectorWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFailedFiles},
	}
	passed := runner.RunV4([]string{testV4WithDocumentSelectorChart})
	assert.False(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithFakeK8sClient(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithFakeK8sClientChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithPostRenderer(t *testing.T) {
	setPostRendererPluginEnv(t)
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithPostRendererChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkWithSchemaValidation(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSchemaChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkPackagedChartWithExternalUnittest(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testExternalSubTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithPackagedSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOkPackagedChart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testExternalTestFiles},
		ValuesFiles: []string{testExternalValuesFiles},
	}
	passed := runner.RunV4([]string{testV4WithPackagedChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV4RunnerOk_With_FailFast_NoPanic(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	cases := []struct {
		chartPath []string
		failFast  bool
	}{
		{
			chartPath: []string{testV4WithFailingTemplateChart},
			failFast:  true,
		},
		{
			chartPath: []string{testV4WithFailingTemplateChart},
			failFast:  false,
		},
		{
			chartPath: []string{testV4InvalidBasicChart},
			failFast:  true,
		},
		{
			chartPath: []string{testV4InvalidBasicChart},
			failFast:  false,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("chart %s fail fast: %v", tt.chartPath[0], tt.failFast), func(t *testing.T) {
			runner.Failfast = tt.failFast
			result := runner.RunV4([]string{testV4WithFailingTemplateChart})
			assert.True(t, result)
		})
	}
}

func TestV4RunnerOkWithDocumentSelect(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithDocumentSelectorChart})
	assert.True(t, passed, buffer.String())

	assert.Contains(t, buffer.String(), "Test Suites: 8 passed, 8 total")
	assert.Contains(t, buffer.String(), "Tests:       13 passed, 13 total")
}

func TestV4RunnerOkWithTestSkipped(t *testing.T) {
	suiteDoc := `
suite: test suite with subchart
templates:
  - charts/postgresql/templates/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV4([]string{testV4WithSchemaChart})
	assert.True(t, passed, buffer.String())
}

func TestV4RunnerOkWithSkippedTests_Output(t *testing.T) {
	chart := `
apiVersion: v2
name: basic
version: 0.1.0
`
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
`
	firstTest := `
---
suite: test skip on suite level
templates:
  - deployment.yaml
tests:
  - it: should render deployment
    asserts:
      - exists:
          path: spec.replicas
`
	secondTest := `
---
suite: test skip on suite level
templates:
  - deployment.yaml
skip:
  reason: "This suite is not ready yet"
tests:
  - it: should render deployment
    asserts:
      - exists:
          path: spec.replicas
`

	thirdFailedTest := `
---
suite: test skip on suite level
templates:
  - deployment.yaml
tests:
  - it: should render deployment
    asserts:
      - exists:
          path: spec.notExists
`
	t.Setenv("GOTMPDIR", ".")
	tmp := t.TempDir()
	paths := []string{filepath.Join(tmp, "chart/templates"), filepath.Join(tmp, "chart/tests")}
	for _, path := range paths {
		err := os.MkdirAll(path, 0755)
		assert.NoError(t, err)
	}

	fs := fstest.MapFS{
		"chart/Chart.yaml":                   {Data: []byte(chart)},
		"chart/templates/deployment.yaml":    {Data: []byte(deployment)},
		"chart/tests/deployment_test.yaml":   {Data: []byte(firstTest)},
		"chart/tests/deployment_2_test.yaml": {Data: []byte(secondTest)},
		"chart/tests/deployment_3_test.yaml": {Data: []byte(thirdFailedTest)},
	}

	for path, el := range fs {
		err := os.WriteFile(filepath.Join(tmp, path), el.Data, 0644)
		assert.NoError(t, err)
	}
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	_ = runner.RunV4([]string{filepath.Join(tmp, "chart")})

	assert.Contains(t, buffer.String(), "Test Suites: 1 failed, 1 passed, 1 skipped, 3 total")
	assert.Contains(t, buffer.String(), "- SKIPPED 'should render deployment'")
	assert.Contains(t, buffer.String(), "Tests:       1 failed, 1 passed, 1 skipped, 3 total")
}

func TestV4RunnerOkWithSkippedSuits_Output(t *testing.T) {
	chart := `
apiVersion: v2
name: basic
version: 0.1.0
`
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
`
	firstTest := `
---
suite: should skip one and execute one
templates:
  - deployment.yaml
tests:
  - it: should skip test
    skip:
     reason: "This suite is not ready yet"
    asserts:
      - exists:
          path: metadata.name
  - it: should not skip test
    asserts:
      - exists:
          path: metadata.name
`
	t.Setenv("GOTMPDIR", ".")
	tmp := t.TempDir()
	paths := []string{filepath.Join(tmp, "chart/templates"), filepath.Join(tmp, "chart/tests")}
	for _, path := range paths {
		err := os.MkdirAll(path, 0755)
		assert.NoError(t, err)
	}

	fs := fstest.MapFS{
		"chart/Chart.yaml":                 {Data: []byte(chart)},
		"chart/templates/deployment.yaml":  {Data: []byte(deployment)},
		"chart/tests/deployment_test.yaml": {Data: []byte(firstTest)},
	}

	for path, el := range fs {
		err := os.WriteFile(filepath.Join(tmp, path), el.Data, 0644)
		assert.NoError(t, err)
	}
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	_ = runner.RunV4([]string{filepath.Join(tmp, "chart")})
	assert.Contains(t, buffer.String(), "PASS  should skip one and execute one")
	assert.Contains(t, buffer.String(), "- SKIPPED 'should skip test'")
	assert.Contains(t, buffer.String(), "Tests:       1 passed, 1 skipped, 2 total")
}

func TestV4RunnerOkWithSkippedTestsWhenSubchartDisabledOnCondition(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		TestFiles:    []string{testTestFiles},
		WithSubChart: true,
	}
	passed := runner.RunV4([]string{testV4WithDisabledSubChartOnConditionChart})
	assert.True(t, passed, buffer.String())

	assert.Contains(t, buffer.String(), "Charts:      1 passed, 1 total")
	assert.Contains(t, buffer.String(), "Test Suites: 0 passed, 0 total")
}

func TestV4RunnerOkWithSkippedTestsWhenSubchartDisabledOnTags(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		TestFiles:    []string{testTestFiles},
		WithSubChart: true,
	}
	passed := runner.RunV4([]string{testV4WithDisabledSubChartOnTagsChart})
	assert.True(t, passed, buffer.String())

	assert.Contains(t, buffer.String(), "Charts:      1 passed, 1 total")
	assert.Contains(t, buffer.String(), "Test Suites: 4 passed, 4 total")
}
