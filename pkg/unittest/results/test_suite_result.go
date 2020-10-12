package results

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lrills/helm-unittest/internal/printer"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
)

// TestSuiteResult result return by TestSuite.Run
type TestSuiteResult struct {
	DisplayName      string
	FilePath         string
	Passed           bool
	ExecError        error
	TestsResult      []*TestJobResult
	SnapshotCounting struct {
		Total    uint
		Failed   uint
		Created  uint
		Vanished uint
	}
}

func (tsr TestSuiteResult) printTitle(printer *printer.Printer) {
	var label string
	if tsr.Passed {
		label = printer.SuccessLabel(" PASS ")
	} else {
		label = printer.DangerLabel(" FAIL ")
	}
	var pathToPrint string
	if tsr.FilePath != "" {
		pathToPrint = printer.Faint(filepath.Dir(tsr.FilePath)+string(os.PathSeparator)) +
			filepath.Base(tsr.FilePath)
	}
	name := printer.Highlight(tsr.DisplayName)
	printer.Println(
		fmt.Sprintf("%s %s\t%s", label, name, pathToPrint),
		0,
	)
}

// Print printing the testresult output.
func (tsr TestSuiteResult) Print(printer *printer.Printer, verbosity int) {
	tsr.printTitle(printer)
	if tsr.ExecError != nil {
		printer.Println(printer.Highlight("- Execution Error: "), 1)
		printer.Println(tsr.ExecError.Error()+"\n", 2)
		return
	}

	for _, result := range tsr.TestsResult {
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
