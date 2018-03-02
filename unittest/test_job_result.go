package unittest

type TestJobResult struct {
	DisplayName         string
	Index               int
	Passed              bool
	ExecError           error
	AssertsResult       []*AssertionResult
	TotalSnapshotCount  uint
	FailedSnapshotCount uint
}

func (tjr TestJobResult) print(printer loggable, verbosity int) {
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
