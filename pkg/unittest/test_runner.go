package unittest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/formatter"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	log "github.com/sirupsen/logrus"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3loader "helm.sh/helm/v3/pkg/chart/loader"
	v3util "helm.sh/helm/v3/pkg/chartutil"
)

const LOG_TEST_RUNNER = "test-runner"

// testUnitCounting stores counting numbers of test unit status
type testUnitCounting struct {
	passed  uint
	failed  uint
	errored uint
	skipped uint
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
	result := failedLabel + erroredLabel
	if counting.skipped > 0 {
		result += fmt.Sprintf(
			"%d passed, %d skipped, %d total",
			counting.passed,
			counting.skipped,
			counting.passed+counting.failed+counting.skipped,
		)
	} else {
		result += fmt.Sprintf(
			"%d passed, %d total",
			counting.passed,
			counting.passed+counting.failed,
		)
	}
	return result
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
		testSuites, err := tr.getV3TestSuitesWithValues(chartPath, chartRoute, chart, nil)
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
		chartPassed := tr.runV3SuitesOfChart(testSuites, chart)

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
	testFilesSet, terr := GetFiles(chartPath, tr.TestFiles, false)
	if terr != nil {
		return nil, terr
	}

	valuesFilesSet, verr := GetFiles("", tr.ValuesFiles, true)
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

// buildMergedValuesForChart builds merged values for a chart to be used for dependency condition evaluation.
// It merges the chart's default values with user-specified values files from the TestRunner.
//
// chart is the chart object for which to build merged values.
// chartPath is the file system path to the chart directory.
//
// It returns the merged values as v3util.Values and an error if any occurred during processing.
func (tr *TestRunner) buildMergedValuesForChart(chart *v3chart.Chart, chartPath string) (v3util.Values, error) {
	base := chart.Values
	if base == nil {
		base = make(map[string]any)
	}

	for _, valuesFile := range tr.ValuesFiles {
		valuesPath := valuesFile
		if !filepath.IsAbs(valuesFile) {
			valuesPath = filepath.Join(chartPath, valuesFile)
		}

		byteArray, err := os.ReadFile(valuesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s: %w", valuesFile, err)
		}

		value := make(map[string]any)
		if err := common.YmlUnmarshal(string(byteArray), &value); err != nil {
			return nil, fmt.Errorf("failed to parse values file %s: %w", valuesFile, err)
		}

		base = v3util.MergeTables(value, base)
	}

	return v3util.Values(base), nil
}

// evaluateConditionPath evaluates a YAML path (e.g., "postgresql.enabled" or "subchart.component.enabled")
// against the provided values and returns the boolean result.
//
// conditionPath is a dot-separated path like "postgresql.enabled"
// values are the merged values to evaluate against
//
// Returns true if the path resolves to a boolean true value, or if the path doesn't exist or isn't a boolean
func evaluateConditionPath(conditionPath string, values v3util.Values) bool {
	if conditionPath == "" {
		return true
	}

	pathParts := strings.Split(conditionPath, ".")
	currentMap := map[string]any(values)

	for i, part := range pathParts {
		if currentMap == nil {
			return true
		}

		next, exists := currentMap[part]
		if !exists {
			return true
		}

		if i == len(pathParts)-1 {
			boolVal, ok := next.(bool)
			if !ok {
				return true
			}
			return boolVal
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return true
		}
		currentMap = nextMap
	}

	return true
}

// evaluateTagsCondition evaluates whether a subchart is enabled based on its tags.
// According to Helm's logic, a subchart is enabled if ANY of its tags evaluate to true (OR logic).
// If no tags are specified, or if the tags field doesn't exist in values, the subchart is enabled by default.
//
// tags is the list of tag names from the dependency declaration
// values are the merged values to evaluate against
//
// Returns true if any tag is true, or if no tags are specified/found in values
func evaluateTagsCondition(tags []string, values v3util.Values) bool {
	if len(tags) == 0 {
		return true
	}

	// Look for the "tags" map in values
	tagsMap, exists := values["tags"]
	if !exists {
		return true
	}

	tagsValues, ok := tagsMap.(map[string]any)
	if !ok {
		return true
	}

	// OR logic: if ANY tag is true, the subchart is enabled
	for _, tag := range tags {
		tagValue, exists := tagsValues[tag]
		if !exists {
			continue
		}

		// Check if the tag value is a boolean true
		if boolVal, ok := tagValue.(bool); ok && boolVal {
			return true
		}
	}

	// All tags are either false or don't exist, subchart is disabled
	return false
}

// getDependencyMetadata finds the Dependency metadata from the parent chart's Chart.yaml
// that matches the given subchart. This is needed because the subchart Chart object doesn't
// contain the condition field - that's only in the parent's dependency declaration.
//
// parentChart is the parent chart containing the dependency declaration
// subchartName is the name of the subchart (which might be an alias)
//
// Returns the Dependency metadata if found, nil otherwise
func getDependencyMetadata(parentChart *v3chart.Chart, subchartName string) *v3chart.Dependency {
	if parentChart.Metadata == nil || parentChart.Metadata.Dependencies == nil {
		return nil
	}

	for _, dep := range parentChart.Metadata.Dependencies {
		// Match by alias if set, otherwise match by name
		matchName := dep.Alias
		if matchName == "" {
			matchName = dep.Name
		}

		if matchName == subchartName {
			return dep
		}
	}

	return nil
}

// isSubchartEnabled evaluates whether a subchart dependency is enabled based on its condition and tags fields
// declared in the parent chart's Chart.yaml dependencies section.
//
// According to Helm's logic:
// 1. If a condition is specified, it takes precedence and determines enablement
// 2. If no condition is specified, tags are evaluated (OR logic: any tag true means enabled)
// 3. If neither condition nor tags are specified, the subchart is enabled by default
//
// parentChart is the parent chart containing the dependency declaration with the condition/tags fields
// subchart is the loaded subchart Chart object
// values are the merged values to evaluate the condition/tags against
//
// It returns true if the subchart should be enabled, false otherwise.
func (tr *TestRunner) isSubchartEnabled(parentChart *v3chart.Chart, subchart *v3chart.Chart, values v3util.Values) bool {
	if subchart.Metadata == nil {
		return true
	}

	subchartName := subchart.Metadata.Name

	// Get the dependency metadata from parent chart's Chart.yaml
	depMetadata := getDependencyMetadata(parentChart, subchartName)
	if depMetadata == nil {
		// No dependency metadata found (shouldn't happen), default to enabled
		return true
	}

	// Priority 1: Check the condition field (takes precedence over tags)
	if depMetadata.Condition != "" {
		return evaluateConditionPath(depMetadata.Condition, values)
	}

	// Priority 2: Check tags field (only if no condition is specified)
	if len(depMetadata.Tags) > 0 {
		return evaluateTagsCondition(depMetadata.Tags, values)
	}

	// No condition or tags specified, subchart is enabled by default
	return true
}

// getV3TestSuitesWithValues retrieves the list of test suites for the given chart and its dependencies (if WithSubChart is true).
// It recursively calls itself for each subchart dependency. This function accepts pre-computed merged values to avoid
// recomputing them for each subchart, and to ensure values file paths are resolved relative to the root chart.
//
// chartPath is the file system path to the chart directory.
// chartRoute is the route/path to the chart within the chart repository.
// chart is the chart object representing the chart being processed.
// mergedValues are the pre-computed merged values (nil means compute them from chartPath).
//
// It returns a slice of TestSuite pointers and an error if any occurred during processing.
func (tr *TestRunner) getV3TestSuitesWithValues(chartPath, chartRoute string, chart *v3chart.Chart, mergedValues v3util.Values) ([]*TestSuite, error) {
	resultSuites, err := tr.getTestSuites(chartPath, chartRoute)
	if err != nil {
		return nil, err
	}

	if !tr.WithSubChart {
		return resultSuites, nil
	}

	// Build merged values for condition evaluation if not provided
	if mergedValues == nil {
		mergedValues, err = tr.buildMergedValuesForChart(chart, chartPath)
		if err != nil {
			log.WithField(LOG_TEST_RUNNER, "get-v3-test-suites").
				Warnf("Failed to merge values for condition evaluation: %v. All subchart tests will be included.", err)
		}
	}

	for _, subchart := range chart.Dependencies() {
		if mergedValues != nil && !tr.isSubchartEnabled(chart, subchart, mergedValues) {
			log.WithField(LOG_TEST_RUNNER, "get-v3-test-suites").
				Debugf("Skipping tests for disabled subchart: %s (from chart: %s)", subchart.Metadata.Name, chart.Name())
			continue
		}

		subchartSuites, err := tr.getV3TestSuitesWithValues(
			filepath.Join(chartPath, "charts", subchart.Metadata.Name),
			filepath.Join(chartRoute, "charts", subchart.Metadata.Name),
			subchart,
			mergedValues,
		)
		if err != nil {
			log.WithField(LOG_TEST_RUNNER, "get-v3-test-suites").
				Warnf("Failed to get test suites for subchart %s: %v", subchart.Metadata.Name, err)
			continue
		}
		resultSuites = append(resultSuites, subchartSuites...)
	}

	return resultSuites, nil
}

// getV3TestSuites retrieves test suites for the given chart and its dependencies (if WithSubChart is true).
// This is a convenience wrapper that automatically computes merged values.
//
// It returns a slice of TestSuite pointers and an error if any occurred during processing.
func (tr *TestRunner) getV3TestSuites(chartPath, chartRoute string, chart *v3chart.Chart) ([]*TestSuite, error) {
	return tr.getV3TestSuitesWithValues(chartPath, chartRoute, chart, nil)
}

// runV3SuitesOfChart runs suite files of the chart and print output
func (tr *TestRunner) runV3SuitesOfChart(suites []*TestSuite, chart *v3chart.Chart) bool {
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
		result := suite.RunV3(chart, snapshotCache, tr.Failfast, tr.RenderPath, &results.TestSuiteResult{})
		chartPassed = chartPassed && result.Passed
		tr.handleSuiteResult(result)
		tr.testResults = append(tr.testResults, result)

		_, storeErr := snapshotCache.StoreToFileIfNeeded()
		if storeErr != nil {
			tr.handleSuiteResult(&results.TestSuiteResult{
				FilePath:  suite.SnapshotFileUrl(),
				ExecError: storeErr,
			})
			chartPassed = false
		}

		if !chartPassed && result.FailFast {
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
		if testsResult == nil {
			if tr.Failfast {
				log.WithField("test-runner", "handle-suite-result").Debug("--failfast skip test")
			}
			continue
		}
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
		tr.Printer.Highlight("%s", chartName),
		tr.Printer.Faint("%s", path),
	)
	tr.Printer.Println(header, 0)
}

// printErroredChartHeader if chart has exexution error print header with error
func (tr *TestRunner) printErroredChartHeader(err error) {
	headerFormat := `
### ` + tr.Printer.Danger("%s", "Error: ") + ` %s
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
			tr.Printer.Faint("%s", " Check changes and use `-u` to update snapshot.")

		tr.Printer.Println(fmt.Sprintf(snapshotFormat, summary), 0)
	}
}

// countSuite count suite status and snapshot status
func (tr *TestRunner) countSuite(suite *results.TestSuiteResult) {
	if suite.Skipped {
		tr.suiteCounting.skipped++
	} else if suite.Passed {
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

// countTest count test status
func (tr *TestRunner) countTest(test *results.TestJobResult) {
	if test.Passed {
		tr.testCounting.passed++
	} else if test.Skipped {
		tr.testCounting.skipped++
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
		defer func() {
			werr := writer.Close()
			if werr != nil {
				log.WithField(LOG_TEST_RUNNER, "write-test-output").Errorf("Error closing output file: %s", werr)
			}
		}()

		//
		jerr := tr.Formatter.WriteTestOutput(tr.testResults, true, writer)
		if jerr != nil {
			return jerr
		}
	}

	return nil
}
