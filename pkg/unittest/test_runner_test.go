package unittest_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

var sectionBeginPattern = regexp.MustCompile("( PASS | FAIL |\n*###|\n*Charts:|\n*Snapshot Summary:)")
var timePattern = regexp.MustCompile(`(Time:\s+)(?:[\d\.]+)(s|ms|\xB5s)`) // B5 = micron for microseconds

func makeOutputSnapshotable(originalOutput string) []interface{} {
	output := strings.ReplaceAll(originalOutput, "\\", "/")
	timeAgnosticOutput := timePattern.ReplaceAllString(output, "${1}XX.XXXms")

	sectionBeggingLocs := sectionBeginPattern.FindAllStringIndex(timeAgnosticOutput, -1)
	sections := make([]string, len(sectionBeggingLocs))

	suiteBeginIdx := -1
	for sectionIdx := 0; sectionIdx < len(sections); sectionIdx++ {
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

	sectionsToRetrun := make([]interface{}, len(sections))
	for idx, section := range sections {
		sectionsToRetrun[idx] = section
	}
	return sectionsToRetrun
}

func TestV3RunnerInvalidChartDirFailfast(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		Failfast:  true,
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testTestFiles})
	assert.False(t, passed, buffer.String())
}

func TestV3RunnerInvalidTestSuiteFailfast(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		Strict:    false,
		Failfast:  true,
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3InvalidBasicChart})
	assert.False(t, passed, buffer.String())
}

func TestV3RunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithSubSubChartsPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		WithSubChart: true,
		Printer:      printer.NewPrinter(buffer, nil),
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithSubSubFolderChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFailingTemplatePassedTest(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithFailingTemplateChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithOverrideValuesPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testTestFiles},
		ValuesFiles: []string{testValuesFiles},
	}
	passed := runner.RunV3([]string{testV3BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithAbsoluteOverrideValuesPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	fullPath, _ := filepath.Abs(testValuesFiles)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testTestFiles},
		ValuesFiles: []string{fullPath},
	}
	passed := runner.RunV3([]string{testV3BasicChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFailedFiles},
	}
	passed := runner.RunV3([]string{testV3BasicChart})
	assert.False(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithSubSubfolder(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithSubFolderChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		WithSubChart: true,
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		WithSubChart: false,
		TestFiles:    []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkGlobalDoubleWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3GlobalDoubleChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFiles(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithFilesChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFullsnapshot(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3FullSnapshotChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithRenderedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:        printer.NewPrinter(buffer, nil),
		ChartTestsPath: "tests-chart",
	}
	passed := runner.RunV3([]string{testV3WithHelmTestsChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithDocumentSelector(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithDocumentSelectorChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithDocumentSelectorWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFailedFiles},
	}
	passed := runner.RunV3([]string{testV3WithDocumentSelectorChart})
	assert.False(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithFakeK8sClient(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithFakeK8sClientChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithPostRenderer(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithPostRendererChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkWithSchemaValidation(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithSchemaChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkPackagedChartWithExternalUnittest(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testExternalSubTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithPackagedSubChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOkPackagedChart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:     printer.NewPrinter(buffer, nil),
		TestFiles:   []string{testExternalTestFiles},
		ValuesFiles: []string{testExternalValuesFiles},
	}
	passed := runner.RunV3([]string{testV3WithPackagedChart})
	assert.True(t, passed, buffer.String())
	cupaloy.SnapshotT(t, makeOutputSnapshotable(buffer.String())...)
}

func TestV3RunnerOk_With_FailFast_NoPanic(t *testing.T) {
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
			chartPath: []string{testV3WithFailingTemplateChart},
			failFast:  true,
		},
		{
			chartPath: []string{testV3WithFailingTemplateChart},
			failFast:  false,
		},
		{
			chartPath: []string{testV3InvalidBasicChart},
			failFast:  true,
		},
		{
			chartPath: []string{testV3InvalidBasicChart},
			failFast:  false,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("chart %s fail fast: %v", tt.chartPath[0], tt.failFast), func(t *testing.T) {
			runner.Failfast = tt.failFast
			result := runner.RunV3([]string{testV3WithFailingTemplateChart})
			assert.True(t, result)
		})
	}
}

func TestV3RunnerOkWithDocumentSelect(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testTestFiles},
	}
	passed := runner.RunV3([]string{testV3WithDocumentSelectorChart})
	assert.True(t, passed, buffer.String())

	assert.Contains(t, buffer.String(), "Test Suites: 8 passed, 8 total")
	assert.Contains(t, buffer.String(), "Tests:       13 passed, 13 total")
}

func TestV3RunnerOkWithTestSkipped(t *testing.T) {
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
	passed := runner.RunV3([]string{testV3WithSchemaChart})
	assert.True(t, passed, buffer.String())
}

func TestV3RunnerOkWithSkippedTests_Output(t *testing.T) {
	if runtime.GOOS == "windows" {
		// https://github.com/golang/go/issues/69159
		t.Skip("This test is not working on Windows. Please try to fix it if you have a Windows machine.")
	}
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
	_ = runner.RunV3([]string{filepath.Join(tmp, "chart")})

	assert.Contains(t, buffer.String(), "Test Suites: 1 failed, 1 passed, 1 skipped, 3 total")
	assert.Contains(t, buffer.String(), "- SKIPPED 'should render deployment'")
	assert.Contains(t, buffer.String(), "Tests:       1 failed, 1 passed, 1 skipped, 3 total")
}

func TestV3RunnerOkWithSkippedSuits_Output(t *testing.T) {
	if runtime.GOOS == "windows" {
		// https://github.com/golang/go/issues/69159
		t.Skip("This test is not working on Windows. Please try to fix it if you have a Windows machine.")
	}
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
	_ = runner.RunV3([]string{filepath.Join(tmp, "chart")})
	assert.Contains(t, buffer.String(), "PASS  should skip one and execute one")
	assert.Contains(t, buffer.String(), "- SKIPPED 'should skip test'")
	assert.Contains(t, buffer.String(), "Tests:       1 passed, 1 skipped, 2 total")
}
