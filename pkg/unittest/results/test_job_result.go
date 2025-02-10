package results

import (
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
)

// TestJobResult result return by TestJob.Run
type TestJobResult struct {
	DisplayName   string
	Index         int
	Passed        bool
	ExecError     error
	AssertsResult []*AssertionResult
	Duration      time.Duration
}

func (tjr TestJobResult) print(printer *printer.Printer, verbosity int) {
	if tjr.Passed {
		return
	}

	if tjr.ExecError != nil {
		printer.Println(printer.Highlight("- %s", tjr.DisplayName), 1)
		printer.Println(printer.Highlight("Error: %s\n", tjr.ExecError.Error()), 2)
		return
	}

	printer.Println(printer.Danger("- %s\n", tjr.DisplayName), 1)
	for _, assertResult := range tjr.AssertsResult {
		assertResult.print(printer, verbosity)
	}
}

// Stringify writing the object to a customized formatted string.
func (tjr TestJobResult) Stringify() string {
	content := ""
	if tjr.ExecError != nil {
		content += tjr.ExecError.Error() + "\n"
	}

	for _, assertResult := range tjr.AssertsResult {
		content += assertResult.stringify()
	}

	return content
}
