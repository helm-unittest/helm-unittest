package unittest_test

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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

// summaryCountLines returns just the summary counting lines (Charts/Test Suites/
// Tests/Snapshot) so parallel and sequential runs can be compared regardless of
// timing noise in the footer.
func summaryCountLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Charts:") ||
			strings.HasPrefix(trimmed, "Test Suites:") ||
			strings.HasPrefix(trimmed, "Tests:") ||
			strings.HasPrefix(trimmed, "Snapshot:") {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func TestV4RunnerParallelSameOutcomeAsSequential(t *testing.T) {
	seqBuffer := new(bytes.Buffer)
	seqRunner := TestRunner{
		Printer:   printer.NewPrinter(seqBuffer, nil),
		TestFiles: []string{testTestFiles},
	}
	seqPassed := seqRunner.RunV4([]string{testV4BasicChart})

	parBuffer := new(bytes.Buffer)
	parRunner := TestRunner{
		Printer:   printer.NewPrinter(parBuffer, nil),
		TestFiles: []string{testTestFiles},
		Parallel:  true,
	}
	parPassed := parRunner.RunV4([]string{testV4BasicChart})

	assert.Equal(t, seqPassed, parPassed)
	assert.Equal(t, summaryCountLines(seqBuffer.String()), summaryCountLines(parBuffer.String()))
}

func TestV4RunnerParallelDeterministicOutput(t *testing.T) {
	run := func(parallel bool) string {
		buffer := new(bytes.Buffer)
		runner := TestRunner{
			Printer:   printer.NewPrinter(buffer, nil),
			TestFiles: []string{testTestFiles},
			Parallel:  parallel,
		}
		runner.RunV4([]string{testV4BasicChart})
		return timePattern.ReplaceAllString(buffer.String(), "${1}XX.XXXms")
	}

	sequential := run(false)
	parallelFirst := run(true)
	parallelSecond := run(true)

	assert.Equal(t, parallelFirst, parallelSecond, "parallel output must be stable across runs")
	assert.Equal(t, sequential, parallelFirst, "parallel output must match sequential order")
}

func TestV4RunnerParallelMultiSuiteSharedSnapshot(t *testing.T) {
	// Copy the fixture so the update run writes into a throwaway location.
	copyChart := func() string {
		t.Setenv("GOTMPDIR", ".")
		dir := filepath.Join(t.TempDir(), "chart")
		if err := os.CopyFS(dir, os.DirFS(testV4ParallelMultiSuiteChart)); err != nil {
			t.Fatalf("failed to copy fixture: %v", err)
		}
		return dir
	}

	parChart := copyChart()
	parBuffer := new(bytes.Buffer)
	parRunner := TestRunner{
		Printer:        printer.NewPrinter(parBuffer, nil),
		TestFiles:      []string{testTestFiles},
		Parallel:       true,
		UpdateSnapshot: true,
	}
	assert.True(t, parRunner.RunV4([]string{parChart}), parBuffer.String())

	// The suites share one .snap file; grouping keeps them on a single goroutine
	// so the file must be byte-identical to a sequential update and valid YAML.
	seqChart := copyChart()
	seqBuffer := new(bytes.Buffer)
	seqRunner := TestRunner{
		Printer:        printer.NewPrinter(seqBuffer, nil),
		TestFiles:      []string{testTestFiles},
		UpdateSnapshot: true,
	}
	assert.True(t, seqRunner.RunV4([]string{seqChart}), seqBuffer.String())

	snapPath := filepath.Join("tests", "__snapshot__", "parallel_test.yaml.snap")
	parSnap, err := os.ReadFile(filepath.Join(parChart, snapPath))
	assert.NoError(t, err)
	seqSnap, err := os.ReadFile(filepath.Join(seqChart, snapPath))
	assert.NoError(t, err)
	assert.Equal(t, string(seqSnap), string(parSnap), "parallel snapshot file must match sequential")

	var parsed map[string]any
	assert.NoError(t, common.YmlUnmarshal(string(parSnap), &parsed), "snapshot file must be valid YAML")

	// Re-running in parallel against the freshly written snapshot must pass.
	rerunBuffer := new(bytes.Buffer)
	rerunRunner := TestRunner{
		Printer:   printer.NewPrinter(rerunBuffer, nil),
		TestFiles: []string{testTestFiles},
		Parallel:  true,
	}
	assert.True(t, rerunRunner.RunV4([]string{parChart}), rerunBuffer.String())
}

// copyMultiSuiteChart copies the multi-suite fixture into a throwaway directory so
// runs that write snapshots never touch the committed fixture.
func copyMultiSuiteChart(t *testing.T) string {
	t.Helper()
	t.Setenv("GOTMPDIR", ".")
	dir := filepath.Join(t.TempDir(), "chart")
	if err := os.CopyFS(dir, os.DirFS(testV4ParallelMultiSuiteChart)); err != nil {
		t.Fatalf("failed to copy fixture: %v", err)
	}
	return dir
}

func multiSuiteSnapshotPath(chartDir string) string {
	return filepath.Join(chartDir, "tests", "__snapshot__", "parallel_test.yaml.snap")
}

func runMultiSuiteChart(chartDir string, parallel, update bool) (bool, string) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:        printer.NewPrinter(buffer, nil),
		TestFiles:      []string{testTestFiles},
		Parallel:       parallel,
		UpdateSnapshot: update,
	}
	return runner.RunV4([]string{chartDir}), buffer.String()
}

// All suites of one test file write to the same .snap file. Every suite's
// snapshots must survive, instead of the last suite overwriting the file.
func TestV4RunnerSharedSnapshotFileKeepsEverySuite(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		t.Run(fmt.Sprintf("parallel=%v", parallel), func(t *testing.T) {
			chartDir := copyMultiSuiteChart(t)
			passed, output := runMultiSuiteChart(chartDir, parallel, true)
			assert.True(t, passed, output)

			snap, err := os.ReadFile(multiSuiteSnapshotPath(chartDir))
			assert.NoError(t, err)

			var parsed map[string]any
			assert.NoError(t, common.YmlUnmarshal(string(snap), &parsed))
			assert.ElementsMatch(t,
				[]string{"renders configmap", "renders service", "renders configmap with override"},
				slices.Collect(maps.Keys(parsed)),
			)
		})
	}
}

// Once written, re-running without -u must compare against the stored snapshots
// and leave the file untouched.
func TestV4RunnerSharedSnapshotFileIsStableAcrossRuns(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		t.Run(fmt.Sprintf("parallel=%v", parallel), func(t *testing.T) {
			chartDir := copyMultiSuiteChart(t)
			passed, output := runMultiSuiteChart(chartDir, parallel, true)
			assert.True(t, passed, output)

			written, err := os.ReadFile(multiSuiteSnapshotPath(chartDir))
			assert.NoError(t, err)

			passed, output = runMultiSuiteChart(chartDir, parallel, false)
			assert.True(t, passed, output)

			after, err := os.ReadFile(multiSuiteSnapshotPath(chartDir))
			assert.NoError(t, err)
			assert.Equal(t, string(written), string(after), "a verifying run must not rewrite the snapshot file")
		})
	}
}

// A changed template must be detected for every suite in the shared file, not
// only for the suite that happened to write the file last.
func TestV4RunnerSharedSnapshotFileDetectsChangePerSuite(t *testing.T) {
	chartDir := copyMultiSuiteChart(t)
	passed, output := runMultiSuiteChart(chartDir, false, true)
	assert.True(t, passed, output)

	// Change the service template, which only the second suite renders.
	servicePath := filepath.Join(chartDir, "templates", "service.yaml")
	original, err := os.ReadFile(servicePath)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(servicePath, bytes.ReplaceAll(original, []byte("port: 80"), []byte("port: 8080")), 0644))

	passed, output = runMultiSuiteChart(chartDir, false, false)
	assert.False(t, passed, "changing a template must fail the suite that snapshots it")
	assert.Contains(t, output, "service suite")
}

// Each suite reports only its own snapshots, so sharing one cache must not make
// the summary count the same snapshot several times.
func TestV4RunnerSharedSnapshotFileCountsEachSuiteOnce(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		t.Run(fmt.Sprintf("parallel=%v", parallel), func(t *testing.T) {
			chartDir := copyMultiSuiteChart(t)
			_, output := runMultiSuiteChart(chartDir, parallel, true)
			assert.Contains(t, output, "Test Suites: 3 passed, 3 total")
			assert.Contains(t, output, "Snapshot:    3 passed, 3 total")
		})
	}
}

// Entries of tests that no longer exist must still be pruned from a shared file.
func TestV4RunnerSharedSnapshotFileRemovesStaleEntries(t *testing.T) {
	chartDir := copyMultiSuiteChart(t)
	passed, output := runMultiSuiteChart(chartDir, false, true)
	assert.True(t, passed, output)

	snapPath := multiSuiteSnapshotPath(chartDir)
	written, err := os.ReadFile(snapPath)
	assert.NoError(t, err)
	stale := string(written) + "renders something removed long ago:\n  1: |\n    stale: true\n"
	assert.NoError(t, os.WriteFile(snapPath, []byte(stale), 0644))

	passed, output = runMultiSuiteChart(chartDir, false, true)
	assert.True(t, passed, output)

	after, err := os.ReadFile(snapPath)
	assert.NoError(t, err)
	assert.NotContains(t, string(after), "renders something removed long ago")
	assert.Equal(t, string(written), string(after))
}

// The fixture ships its snapshots, so the runs above compare against known content
// instead of silently regenerating it (helm-unittest#246). Guard that the committed
// snapshot stays in sync with the fixture and is never rewritten by a verifying run.
func TestV4RunnerMultiSuiteFixtureShipsItsSnapshot(t *testing.T) {
	snapPath := multiSuiteSnapshotPath(testV4ParallelMultiSuiteChart)
	committed, err := os.ReadFile(snapPath)
	assert.NoError(t, err, "the fixture must ship its snapshot file")

	passed, output := runMultiSuiteChart(testV4ParallelMultiSuiteChart, true, false)
	assert.True(t, passed, output)

	after, err := os.ReadFile(snapPath)
	assert.NoError(t, err)
	assert.Equal(t, string(committed), string(after), "a verifying run must not rewrite the committed snapshot")
}

func TestV4RunnerParallelFailfastStopsScheduling(t *testing.T) {
	countFailedSuites := func(output string) int {
		return strings.Count(output, " FAIL ")
	}

	// Work on a throwaway copy so the failing suites (which may rewrite snapshots)
	// never touch the committed fixture.
	chartDir := filepath.Join(t.TempDir(), "chart")
	if err := os.CopyFS(chartDir, os.DirFS(testV4BasicChart)); err != nil {
		t.Fatalf("failed to copy fixture: %v", err)
	}

	// Without failfast, every failing suite is reported.
	fullBuffer := new(bytes.Buffer)
	fullRunner := TestRunner{
		Printer:   printer.NewPrinter(fullBuffer, nil),
		TestFiles: []string{testTestFailedFiles},
		Parallel:  true,
	}
	assert.False(t, fullRunner.RunV4([]string{chartDir}), fullBuffer.String())
	fullCount := countFailedSuites(fullBuffer.String())
	assert.Greater(t, fullCount, 1)

	// With failfast and a single worker, the first failure prevents the remaining
	// groups from being scheduled, so fewer suites are reported.
	ffBuffer := new(bytes.Buffer)
	ffRunner := TestRunner{
		Printer:    printer.NewPrinter(ffBuffer, nil),
		TestFiles:  []string{testTestFailedFiles},
		Parallel:   true,
		Failfast:   true,
		MaxWorkers: 1,
	}
	assert.False(t, ffRunner.RunV4([]string{chartDir}), ffBuffer.String())
	assert.Less(t, countFailedSuites(ffBuffer.String()), fullCount, "failfast must skip unstarted groups")
}
