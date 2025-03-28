package results

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
)

// TestSuiteResult result return by TestSuite.Run
type TestSuiteResult struct {
	DisplayName      string
	FilePath         string
	Passed           bool
	Skipped          bool
	FailFast         bool
	ExecError        error
	TestsResult      []*TestJobResult
	SnapshotCounting struct {
		Total    uint
		Failed   uint
		Created  uint
		Vanished uint
	}
}

// printTitle print the title of the test suite.
func (tsr *TestSuiteResult) printTitle(printer *printer.Printer) {
	var label string
	if tsr.Skipped {
		label = printer.WarningLabel(" SKIP ")
	} else if tsr.Passed {
		label = printer.SuccessLabel(" PASS ")
	} else {
		label = printer.DangerLabel(" FAIL ")
	}
	var pathToPrint string
	if tsr.FilePath != "" {
		pathToPrint = printer.Faint("%s", filepath.ToSlash(filepath.Dir(tsr.FilePath)+string(os.PathSeparator))) +
			filepath.Base(tsr.FilePath)
	}
	name := printer.Highlight("%s", tsr.DisplayName)
	printer.Println(
		fmt.Sprintf("%s %s\t%s", label, name, pathToPrint),
		0,
	)
}

// Print printing the test result output.
func (tsr *TestSuiteResult) Print(printer *printer.Printer, verbosity int) {
	tsr.printTitle(printer)
	if tsr.ExecError != nil {
		printer.Println(printer.Highlight("%s", "- Execution Error: "), 1)
		printer.Println(tsr.ExecError.Error()+"\n", 2)
		return
	}

	for _, result := range tsr.TestsResult {
		if result == nil {
			continue
		}
		result.print(printer, verbosity)
	}
}

// CountSnapshot counting the snaphots.
func (tsr *TestSuiteResult) CountSnapshot(cache *snapshot.Cache) {
	tsr.SnapshotCounting.Created = cache.InsertedCount()
	tsr.SnapshotCounting.Failed = cache.FailedCount()
	tsr.SnapshotCounting.Total = cache.CurrentCount()
	tsr.SnapshotCounting.Vanished = cache.VanishedCount()
}

// CalculateTestSuiteDuration to calculate the total duration of the testsuite.
func (tsr *TestSuiteResult) CalculateTestSuiteDuration() time.Duration {
	var totalDuration time.Duration
	for _, test := range tsr.TestsResult {
		totalDuration += test.Duration
	}
	return totalDuration
}
