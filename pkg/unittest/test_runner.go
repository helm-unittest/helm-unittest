package unittest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/printer"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"

	"github.com/yargevad/filepathx"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3loader "helm.sh/helm/v3/pkg/chart/loader"
)

// getFiles retrieves a list of files matching the given file patterns.
// If chartPath is provided, the patterns are treated as relative to the chartPath.
// If setAbsolute is true, the returned file paths are converted to absolute paths.
//
// chartPath is the base directory to search for files (can be empty).
// filePatterns is a slice of file patterns to match (e.g., "tests/*".yaml).
// setAbsolute indicates whether to return absolute file paths or relative paths.
//
// It returns a slice of file paths and an error if any occurred during processing.
func getFiles(chartPath string, filePatterns []string, setAbsolute bool) ([]string, error) {
	log.WithField("test-runner", "getFiles").Debugln("filepatterns:", filePatterns)
	var filesSet []string
	basePath := chartPath + "/" // Prepend chartPath with slash

	for _, pattern := range filePatterns {
		if filepath.IsAbs(pattern) {
			filesSet = append(filesSet, pattern) // Append absolute paths directly
		} else {
			var filePath string
			if strings.Contains(pattern, basePath) {
				filePath = pattern
			} else {
				filePath = filepath.Join(basePath, pattern)
			}
			files, err := filepathx.Glob(filePath)
			if err != nil {
				return nil, err
			}
			filesSet = append(filesSet, files...) // Append all files (relative)
		}
	}

	if setAbsolute {
		// If setAbsolute is true, convert the file paths to absolute paths
		for i, filePath := range filesSet {
			if !filepath.IsAbs(filePath) {
				absPath, _ := filepath.Abs(filePath)
				filesSet[i] = absPath
			}
		}
	}

	log.WithField("test-runner", "getFiles").Debugln("chartpath:", chartPath, ". fileset:", filesSet)
	return filesSet, nil
}

// testUnitCounting stores counting numbers of test unit status
type testUnitCounting struct {
	passed  uint
	failed  uint
	errored uint
}

// sprint returns string of counting result
func (counting testUnitCounting) sprint(printer *printer.Printer) string {
	var failedLabel string
	if counting.failed > 0 {
		failedLabel = printer.Danger("%d failed, ", counting.failed)
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

// testUnitCountingWithSnapshotFailed store testUnitCounting with snapshotFailed field
type testUnitCountingWithSnapshotFailed struct {
	testUnitCounting
	snapshotFailed uint
}

// totalSnapshotCounting store testUnitCounting with snapshotFailed field
type totalSnapshotCounting struct {
	testUnitCounting
	created  uint
	vanished uint
}

// TestRunner stores basic settings and testing status for running all tests
type TestRunner struct {
	Printer          *printer.Printer
	Formatter        formatter.Formatter
	UpdateSnapshot   bool
	WithSubChart     bool
	Strict           bool
	Failfast         bool
	TestFiles        []string
	ChartTestsPath   string
	ValuesFiles      []string
	OutputFile       string
	RenderPath       string
	suiteCounting    testUnitCountingWithSnapshotFailed
	testCounting     testUnitCounting
	chartCounting    testUnitCounting
	snapshotCounting totalSnapshotCounting
	testResults      []*results.TestSuiteResult
}

// RunV3 test suites in chart in ChartPaths.
func (tr *TestRunner) RunV3(ChartPaths []string) bool {
	allPassed := true
	start := time.Now()
	for _, chartPath := range ChartPaths {
		chart, err := v3loader.Load(chartPath)
		if err != nil {
			tr.printErroredChartHeader(err)
			tr.countChart(false, err)
			allPassed = false
			if tr.Failfast {
				break
			}
			continue
		}
		chartRoute := chart.Name()
		testSuites, err := tr.getV3TestSuites(chartPath, chartRoute, chart)
		if err != nil {
			tr.printErroredChartHeader(err)
			tr.countChart(false, err)
			allPassed = false
			if tr.Failfast {
				break
			}
			continue
		}

		tr.printChartHeader(chart.Name(), chartPath)
		chartPassed := tr.runV3SuitesOfChart(testSuites, chartPath)

		tr.countChart(chartPassed, nil)
		allPassed = allPassed && chartPassed
	}
	err := tr.writeTestOutput()
	if err != nil {
		tr.printErroredChartHeader(err)
	}
	tr.printSnapshotSummary()
	tr.printSummary(time.Since(start))
	return allPassed
}

// getTestSuites retrieves the list of test suites for the given chart.
// It parses test suite files and renders test suite files from the chart's tests path (if specified).
//
// chartPath is the file system path to the chart directory.
// chartRoute is the route/path to the chart within the chart repository.
//
// It returns a slice of _TestSuite structs and an error if any occurred during processing.
func (tr *TestRunner) getTestSuites(chartPath, chartRoute string) ([]*TestSuite, error) {
	testFilesSet, terr := getFiles(chartPath, tr.TestFiles, false)
	if terr != nil {
		return nil, terr
	}

	valuesFilesSet, verr := getFiles("", tr.ValuesFiles, true)
	if verr != nil {
		return nil, verr
	}

	var renderedTestSuites []*TestSuite
	if len(tr.ChartTestsPath) > 0 {
		helmTestsPath := filepath.Join(chartPath, tr.ChartTestsPath)
		// Verify that there is a tests path - in the event of mixed testing environments
		if _, err := os.Stat(helmTestsPath); errors.Is(err, nil) {
			var renderErr error
			renderedTestSuites, renderErr = RenderTestSuiteFiles(helmTestsPath, chartRoute, tr.Strict, valuesFilesSet, nil)
			if renderErr != nil {
				return nil, renderErr
			}
		}
	}

	// Parse each test file and append the resulting test suites to the result slice
	resultSuites := make([]*TestSuite, 0, len(testFilesSet)+len(renderedTestSuites))
	for _, file := range testFilesSet {
		suites, err := ParseTestSuiteFile(file, chartRoute, tr.Strict, valuesFilesSet)
		if err != nil {
			tr.handleSuiteResult(&results.TestSuiteResult{
				FilePath:  file,
				ExecError: err,
			})
			return nil, err
		}
		resultSuites = append(resultSuites, suites...)
	}
	resultSuites = append(resultSuites, renderedTestSuites...)

	return resultSuites, nil
}

// getV3TestSuites retrieves the list of test suites for the given chart and its dependencies (if WithSubChart is true).
// It recursively calls itself for each subchart dependency.
//
// chartPath is the file system path to the chart directory.
// chartRoute is the route/path to the chart within the chart repository.
// chart_ is the chart object representing the chart being processed.
//
// It returns a slice of TestSuite pointers and an error if any occurred during processing.
func (tr *TestRunner) getV3TestSuites(chartPath, chartRoute string, chart *v3chart.Chart) ([]*TestSuite, error) {
	resultSuites, err := tr.getTestSuites(chartPath, chartRoute)
	if err != nil {
		return nil, err
	}

	if tr.WithSubChart {
		for _, subchart := range chart.Dependencies() {
			// Recursively get test suites for the subchart
			subchartSuites, err := tr.getV3TestSuites(
				filepath.Join(chartPath, "charts", subchart.Metadata.Name),
				filepath.Join(chartRoute, "charts", subchart.Metadata.Name),
				subchart,
			)
			if err != nil {
				continue
			}
			resultSuites = append(resultSuites, subchartSuites...)
		}
	}

	return resultSuites, nil
}

// runV3SuitesOfChart runs suite files of the chart and print output
func (tr *TestRunner) runV3SuitesOfChart(suites []*TestSuite, chartPath string) bool {
	chartPassed := true
	for _, suite := range suites {
		snapshotCache, err := snapshot.CreateSnapshotOfSuite(suite.SnapshotFileUrl(), tr.UpdateSnapshot)
		if err != nil {
			tr.handleSuiteResult(&results.TestSuiteResult{
				FilePath:  suite.definitionFile,
				ExecError: err,
			})
			chartPassed = false
			continue
		}

		result := suite.RunV3(chartPath, snapshotCache, tr.Failfast, tr.RenderPath, &results.TestSuiteResult{})
		chartPassed = chartPassed && result.Passed
		tr.handleSuiteResult(result)
		tr.testResults = append(tr.testResults, result)

		snapshotCache.StoreToFileIfNeeded()

		if !chartPassed && tr.Failfast {
			break
		}
	}

	return chartPassed
}

// handleSuiteResult print suite result and count suites and tests status
func (tr *TestRunner) handleSuiteResult(result *results.TestSuiteResult) {
	result.Print(tr.Printer, 0)
	tr.countSuite(result)
	for _, testsResult := range result.TestsResult {
		tr.countTest(testsResult)
	}
}

// printSummary print summary footer
func (tr *TestRunner) printSummary(elapsed time.Duration) {
	summaryFormat := `
Charts:      %s
Test Suites: %s
Tests:       %s
Snapshot:    %s
Time:        %s
`
	tr.Printer.Println(
		fmt.Sprintf(
			summaryFormat,
			tr.chartCounting.sprint(tr.Printer),
			tr.suiteCounting.sprint(tr.Printer),
			tr.testCounting.sprint(tr.Printer),
			tr.snapshotCounting.sprint(tr.Printer),
			elapsed.String(),
		),
		0,
	)

}

// printChartHeader print header before suite result of a chart
func (tr *TestRunner) printChartHeader(chartName, path string) {
	headerFormat := `
### Chart [ %s ] %s
`
	header := fmt.Sprintf(
		headerFormat,
		tr.Printer.Highlight(chartName),
		tr.Printer.Faint(path),
	)
	tr.Printer.Println(header, 0)
}

// printErroredChartHeader if chart has exexution error print header with error
func (tr *TestRunner) printErroredChartHeader(err error) {
	headerFormat := `
### ` + tr.Printer.Danger("Error: ") + ` %s
`
	header := fmt.Sprintf(headerFormat, err)
	tr.Printer.Println(header, 0)
}

// printSnapshotSummary print snapshot summary in footer
func (tr *TestRunner) printSnapshotSummary() {
	if tr.snapshotCounting.failed > 0 {
		snapshotFormat := `
Snapshot Summary: %s`

		summary := tr.Printer.Danger("%d snapshot failed", tr.snapshotCounting.failed) +
			fmt.Sprintf(" in %d test suite.", tr.suiteCounting.snapshotFailed) +
			tr.Printer.Faint(" Check changes and use `-u` to update snapshot.")

		tr.Printer.Println(fmt.Sprintf(snapshotFormat, summary), 0)
	}
}

func (tr *TestRunner) countSuite(suite *results.TestSuiteResult) {
	if suite.Passed {
		tr.suiteCounting.passed++
	} else {
		tr.suiteCounting.failed++
		if suite.ExecError != nil {
			tr.suiteCounting.errored++
		}
		if suite.SnapshotCounting.Failed > 0 {
			tr.suiteCounting.snapshotFailed++
		}
	}
	tr.snapshotCounting.failed += suite.SnapshotCounting.Failed
	tr.snapshotCounting.passed += suite.SnapshotCounting.Total - suite.SnapshotCounting.Failed
	tr.snapshotCounting.created += suite.SnapshotCounting.Created
	tr.snapshotCounting.vanished += suite.SnapshotCounting.Vanished
}

func (tr *TestRunner) countTest(test *results.TestJobResult) {
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

func (tr *TestRunner) writeTestOutput() error {
	// Check if formatter exits to write
	if tr.Formatter != nil {
		// Create outputfile for testsuite
		writer, ferr := os.Create(tr.OutputFile)
		if ferr != nil {
			return ferr
		}
		defer writer.Close()

		//
		jerr := tr.Formatter.WriteTestOutput(tr.testResults, true, writer)
		if jerr != nil {
			return jerr
		}
	}

	return nil
}
