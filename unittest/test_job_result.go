package unittest

import "time"

// TestJobResult result return by TestJob.Run
type TestJobResult struct {
	DisplayName   string
	Index         int
	Passed        bool
	ExecError     error
	AssertsResult []*AssertionResult
	Duration      time.Duration
}

func (tjr TestJobResult) print(printer *Printer, verbosity int) {
	if tjr.Passed {
		return
	}

	if tjr.ExecError != nil {
		printer.println(printer.highlight("- "+tjr.DisplayName), 1)
		printer.println(
			printer.highlight("Error: ")+
				tjr.ExecError.Error()+"\n",
			2,
		)
		return
	}

	printer.println(printer.danger("- "+tjr.DisplayName+"\n"), 1)
	for _, assertResult := range tjr.AssertsResult {
		assertResult.print(printer, verbosity)
	}
}

// ToString writing the object to a customized formatted string.
func (tjr TestJobResult) stringify() string {
	content := ""
	if tjr.ExecError != nil {
		content += tjr.ExecError.Error() + "\n"
	}

	for _, assertResult := range tjr.AssertsResult {
		content += assertResult.stringify()
	}

	return content
}
