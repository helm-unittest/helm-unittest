package results

import (
	"time"

	"github.com/lrills/helm-unittest/internal/printer"
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
		printer.Println(printer.Highlight("- "+tjr.DisplayName), 1)
		printer.Println(
			printer.Highlight("Error: ")+
				tjr.ExecError.Error()+"\n",
			2,
		)
		return
	}

	printer.Println(printer.Danger("- "+tjr.DisplayName+"\n"), 1)
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
