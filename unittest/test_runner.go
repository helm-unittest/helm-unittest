package unittest

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/lrills/helm-unittest/unittest/snapshot"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// getTestSuiteFiles return test files of the chart which matched patterns
func getTestSuiteFiles(chartPath string, patterns []string) ([]string, error) {
	filesSet := map[string]bool{}
	for _, pattern := range patterns {
		files, err := filepath.Glob(filepath.Join(chartPath, pattern))
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			filesSet[file] = true
		}
	}

	resultFiles := make([]string, len(filesSet))
	idx := 0
	for file := range filesSet {
		resultFiles[idx] = file
		idx++
	}
	return resultFiles, nil
}

// testUnitCounting stores counting numbers of test unit status
type testUnitCounting struct {
	passed         uint
	failed         uint
	errored        uint
	snapshotFailed uint
}

// sprint returns string of counting result
func (counting testUnitCounting) sprint(logger loggable) string {
	var failedLabel string
	if counting.failed > 0 {
		failedLabel = logger.danger("%d failed, ", counting.failed)
	}
	var erroredLabel string
	if counting.errored > 0 {
		erroredLabel = fmt.Sprintf("%d errored, ", counting.errored)
	}
	return failedLabel + erroredLabel + fmt.Sprintf(
		"%d passed, %d total",
		counting.passed,
		counting.passed+counting.failed,
	)
}

// testUnitCountingWithSnapshot store testUnitCounting with snapshotFailed field
type testUnitCountingWithSnapshot struct {
	testUnitCounting
	snapshotFailed uint
}

// TestRunner stores basic settings and testing status for running all tests
type TestRunner struct {
	Logger           loggable
	Config           TestConfig
	suiteCounting    testUnitCountingWithSnapshot
	testCounting     testUnitCountingWithSnapshot
	chartCounting    testUnitCounting
	snapshotCounting testUnitCounting
}

// Run test suites in chart in ChartPaths
func (tr *TestRunner) Run(ChartPaths []string) bool {
	allPassed := true
	start := time.Now()
	for _, chartPath := range ChartPaths {
		chart, err := chartutil.Load(chartPath)
		if err != nil {
			tr.printErroredChartHeader(err)
			tr.countChart(false, err)
			allPassed = false
			continue
		}

		suiteFiles, err := getTestSuiteFiles(chartPath, tr.Config.TestFiles)
		if err != nil {
			tr.printErroredChartHeader(err)
			tr.countChart(false, err)
			allPassed = false
			continue
		}

		tr.printChartHeader(chart, chartPath)
		chartPassed := tr.runSuitesOfChart(suiteFiles, chart)

		tr.countChart(chartPassed, nil)
		allPassed = allPassed && chartPassed
	}
	tr.printSnapshotSummary()
	tr.printSummary(time.Now().Sub(start))
	return allPassed
}

// runSuitesOfChart runs suite files of the chart and print output
func (tr *TestRunner) runSuitesOfChart(suiteFiles []string, chart *chart.Chart) bool {
	chartPassed := true
	for _, file := range suiteFiles {
		testSuite, err := ParseTestSuiteFile(file)
		if err != nil {
			tr.handleSuiteResult(&TestSuiteResult{
				FilePath:  file,
				ExecError: err,
			})
		}

		snapshotCache, _ := snapshot.CreateSnapshotOfSuite(file, tr.Config.UpdateSnapshot)
		// TODO: should print warning

		result := testSuite.Run(chart, snapshotCache, &TestSuiteResult{})
		chartPassed = chartPassed && result.Passed
		tr.handleSuiteResult(result)

		if tr.Config.UpdateSnapshot && snapshotCache.Changed() {
			snapshotCache.StoreToFile()
		}
	}

	return chartPassed
}

// handleSuiteResult print suite result and count suites and tests status
func (tr *TestRunner) handleSuiteResult(result *TestSuiteResult) {
	result.print(tr.Logger, 0)
	tr.countSuite(result)
	for _, testsResult := range result.TestsResult {
		tr.countTest(testsResult)
	}
}

//printSummary print summary footer
func (tr *TestRunner) printSummary(elapsed time.Duration) {
	summaryFormat := `
Charts:      %s
Test Suites: %s
Tests:       %s
Snapshot:    %s
Time:        %s
`
	tr.Logger.println(
		fmt.Sprintf(
			summaryFormat,
			tr.chartCounting.sprint(tr.Logger),
			tr.suiteCounting.sprint(tr.Logger),
			tr.testCounting.sprint(tr.Logger),
			tr.snapshotCounting.sprint(tr.Logger),
			elapsed.String(),
		),
		0,
	)

}

// printChartHeader print header before suite result of a chart
func (tr *TestRunner) printChartHeader(chart *chart.Chart, path string) {
	headerFormat := `
### Chart [ %s ] %s
`
	header := fmt.Sprintf(
		headerFormat,
		tr.Logger.highlight(chart.Metadata.Name),
		tr.Logger.faint(path),
	)
	tr.Logger.println(header, 0)
}

// printErroredChartHeader if chart has exexution error print header with error
func (tr *TestRunner) printErroredChartHeader(err error) {
	headerFormat := `
### ` + tr.Logger.danger("Error: ") + ` %s
`
	header := fmt.Sprintf(headerFormat, err)
	tr.Logger.println(header, 0)
}

// printSnapshotSummary print snapshot summary in footer
func (tr *TestRunner) printSnapshotSummary() {
	if tr.snapshotCounting.failed > 0 {
		snapshotFormat := `
		Snapshot Summary: %s`

		summary := tr.Logger.danger("%d snapshot failed", tr.snapshotCounting.failed) +
			fmt.Sprintf(" in %d test suite.", tr.suiteCounting.snapshotFailed) +
			tr.Logger.faint(" Check changes and use `-u` to update snapshot.")

		tr.Logger.println(fmt.Sprintf(snapshotFormat, summary), 0)
	}
}

func (tr *TestRunner) countSuite(suite *TestSuiteResult) {
	if suite.Passed {
		tr.suiteCounting.passed++
	} else {
		tr.suiteCounting.failed++
		if suite.ExecError != nil {
			tr.suiteCounting.errored++
		}
		if suite.HasSnapshotFail {
			tr.suiteCounting.snapshotFailed++
		}
	}
}

func (tr *TestRunner) countTest(test *TestJobResult) {
	if test.Passed {
		tr.testCounting.passed++
	} else {
		tr.testCounting.failed++
		if test.ExecError != nil {
			tr.testCounting.errored++
		}
		if test.FailedSnapshotCount > 0 {
			tr.testCounting.snapshotFailed++
		}
	}
	tr.snapshotCounting.failed += test.FailedSnapshotCount
	tr.snapshotCounting.passed += test.TotalSnapshotCount - test.FailedSnapshotCount
}

func (tr *TestRunner) countChart(passed bool, err error) {
	if passed {
		tr.chartCounting.passed++
	} else {
		tr.chartCounting.failed++
		if err != nil {
			tr.chartCounting.errored++
		}
	}
}
