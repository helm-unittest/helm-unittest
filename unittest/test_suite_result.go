package unittest

import (
	"fmt"
	"os"
	"path/filepath"
)

type TestSuiteResult struct {
	DisplayName string
	FilePath    string
	Passed      bool
	ExecError   error
	TestsResult []*TestJobResult
}

func (tsr TestSuiteResult) print(printer loggable, verbosity int) {
	tsr.printTitle(printer)
	if tsr.ExecError != nil {
		printer.println(printer.highlight("- Execution Error: "), 1)
		printer.println(tsr.ExecError.Error()+"\n", 2)
		return
	}

	for _, result := range tsr.TestsResult {
		result.print(printer, verbosity)
	}
}

func (tsr TestSuiteResult) printTitle(printer loggable) {
	var label string
	if tsr.Passed {
		label = printer.successBackground(" PASS ")
	} else {
		label = printer.dangerBackground(" FAIL ")
	}
	var pathToPrint string
	if tsr.FilePath != "" {
		pathToPrint = filepath.Dir(tsr.FilePath) +
			string(os.PathSeparator) +
			printer.highlight(filepath.Base(tsr.FilePath))
	}
	name := printer.highlight(tsr.DisplayName)
	printer.println(
		fmt.Sprintf("%s %s %s", label, name, pathToPrint),
		0,
	)
}
