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
		failedLabel = logger.danger(fmt.Sprintf("%d failed, ", counting.failed))
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

type TestRunner struct {
	ChartsPath       []string
	suiteCounting    testUnitCountingWithSnapshot
	testCounting     testUnitCountingWithSnapshot
	chartCounting    testUnitCounting
	snapshotCounting testUnitCounting
}

func (tr *TestRunner) Run(logger loggable, config TestConfig) bool {
	allPassed := true
	start := time.Now()
	for _, chartPath := range tr.ChartsPath {
		chart, err := chartutil.Load(chartPath)
		if err != nil {
			tr.printErroredChartHeader(logger, err)
			tr.countChart(false, err)
			allPassed = false
			continue
		}

		suiteFiles, err := getTestSuiteFiles(chartPath, config.TestFiles)
		if err != nil {
			tr.printErroredChartHeader(logger, err)
			tr.countChart(false, err)
			allPassed = false
			continue
		}

		tr.printChartHeader(logger, chart, chartPath)
		chartPassed := tr.runSuitesOfChart(suiteFiles, chart, logger, config)

		tr.countChart(chartPassed, nil)
		allPassed = allPassed && chartPassed
	}
	tr.printSnapshotSummary(logger)
	tr.printSummary(logger, time.Now().Sub(start))
	return allPassed
}

// runSuitesOfChart rusn suite files of the chart and print output
func (tr *TestRunner) runSuitesOfChart(
	suiteFiles []string,
	chart *chart.Chart,
	logger loggable,
	config TestConfig,
) bool {
	chartPassed := true
	for _, file := range suiteFiles {
		testSuite, err := ParseTestSuiteFile(file)
		if err != nil {
			tr.handleSuiteResult(&TestSuiteResult{
				FilePath:  file,
				ExecError: err,
			}, logger)
		}

		snapshotCache, _ := snapshot.CreateSnapshotOfSuite(file, config.UpdateSnapshot)
		// TODO: should print warning

		result := testSuite.Run(chart, snapshotCache, &TestSuiteResult{})
		chartPassed = chartPassed && result.Passed
		tr.handleSuiteResult(result, logger)

		if config.UpdateSnapshot && snapshotCache.Changed() {
			snapshotCache.StoreToFile()
		}
	}

	return chartPassed
}

// handleSuiteResult print suite result and count suites and tests status
func (tr *TestRunner) handleSuiteResult(result *TestSuiteResult, logger loggable) {
	result.print(logger, 0)
	tr.countSuite(result)
	for _, testsResult := range result.TestsResult {
		tr.countTest(testsResult)
	}
}

//printSummary print summary footer
func (tr *TestRunner) printSummary(logger loggable, elapsed time.Duration) {
	summaryFormat := `
Charts:      %s
Test Suites: %s
Tests:       %s
Snapshot:    %s
Time:        %s
`
	logger.println(
		fmt.Sprintf(
			summaryFormat,
			tr.chartCounting.sprint(logger),
			tr.suiteCounting.sprint(logger),
			tr.testCounting.sprint(logger),
			tr.snapshotCounting.sprint(logger),
			elapsed.String(),
		),
		0,
	)

}

// printChartHeader print header before suite result of a chart
func (tr *TestRunner) printChartHeader(logger loggable, chart *chart.Chart, path string) {
	headerFormat := `
### Chart [ %s ] %s
`
	header := fmt.Sprintf(headerFormat, logger.highlight(chart.Metadata.Name), logger.faint(path))
	logger.println(header, 0)
}

// printErroredChartHeader if chart has exexution error print header with error
func (tr *TestRunner) printErroredChartHeader(logger loggable, err error) {
	headerFormat := `
### ` + logger.danger("Error: ") + ` %s
`
	header := fmt.Sprintf(headerFormat, err)
	logger.println(header, 0)
}

// printSnapshotSummary print snapshot summary in footer
func (tr *TestRunner) printSnapshotSummary(logger loggable) {
	if tr.snapshotCounting.failed > 0 {
		snapshotFormat := `
		Snapshot Summary: %s`

		summary := logger.danger(fmt.Sprintf("%d snapshot failed", tr.snapshotCounting.failed)) +
			fmt.Sprintf(" in %d test suite.", tr.suiteCounting.snapshotFailed) +
			logger.faint(" Check changes and use `-u` to update snapshot.")

		logger.println(fmt.Sprintf(snapshotFormat, summary), 0)
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
