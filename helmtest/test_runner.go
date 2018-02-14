package helmtest

import (
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func getTestSuiteFiles(chartPath string) ([]string, error) {
	return filepath.Glob(filepath.Join(chartPath, "tests", "*.yaml"))
}

type testUnitCounting struct {
	passed  uint
	failed  uint
	errored uint
}

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

type TestRunner struct {
	ChartsPath    []string
	suiteCounting testUnitCounting
	testCounting  testUnitCounting
	chartCounting testUnitCounting
}

func (tr *TestRunner) Run(logger loggable) bool {
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

		suiteFiles, err := getTestSuiteFiles(chartPath)
		if err != nil {
			tr.printErroredChartHeader(logger, err)
			tr.countChart(false, err)
			allPassed = false
			continue
		}

		tr.printChartHeader(logger, chart, chartPath)
		chartPassed := tr.runSuites(suiteFiles, chart, logger)
		tr.countChart(chartPassed, nil)
		allPassed = allPassed && chartPassed
	}
	tr.printSummary(logger, time.Now().Sub(start))
	return allPassed
}

func (tr *TestRunner) runSuites(suiteFiles []string, chart *chart.Chart, logger loggable) bool {
	chartPassed := true
	suitesResult := make([]*TestSuiteResult, len(suiteFiles))
	for idx, file := range suiteFiles {
		testSuite, err := ParseTestSuiteFile(file)
		if err != nil {
			suitesResult[idx] = &TestSuiteResult{
				FilePath:  file,
				ExecError: err,
			}
		}
		result := testSuite.Run(chart, &TestSuiteResult{FilePath: file})
		chartPassed = chartPassed && result.Passed
		suitesResult[idx] = result
	}

	for _, suiteResult := range suitesResult {
		suiteResult.print(logger, 0)
		tr.countSuite(suiteResult)
		for _, testsResult := range suiteResult.TestsResult {
			tr.countTest(testsResult)
		}
	}
	return chartPassed
}

func (tr *TestRunner) printSummary(logger loggable, elapsed time.Duration) {
	summaryFormat := `
Charts:      %s
Test Suites: %s
Tests:       %s
Time:        %s
`
	logger.println(
		fmt.Sprintf(
			summaryFormat,
			tr.chartCounting.sprint(logger),
			tr.suiteCounting.sprint(logger),
			tr.testCounting.sprint(logger),
			elapsed.String(),
		),
		0,
	)

}

func (tr *TestRunner) printChartHeader(logger loggable, chart *chart.Chart, path string) {
	headerFormat := `
### Chart [ %s ] %s
`
	header := fmt.Sprintf(headerFormat, logger.highlight(chart.Metadata.Name), path)
	logger.println(header, 0)
}

func (tr *TestRunner) printErroredChartHeader(logger loggable, err error) {
	headerFormat := `
### ` + logger.danger("Error: ") + ` %s
`
	header := fmt.Sprintf(headerFormat, err)
	logger.println(header, 0)
}

func (tr *TestRunner) countSuite(suite *TestSuiteResult) {
	if suite.Passed {
		tr.suiteCounting.passed++
	} else {
		tr.suiteCounting.failed++
		if suite.ExecError != nil {
			tr.suiteCounting.errored++
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
	}
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
