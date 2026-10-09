package results

import (
	"fmt"
	"strings"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
)

// TestJobResult result return by TestJob.Run
type TestJobResult struct {
	DisplayName   string
	Index         int
	Passed        bool
	Skipped       bool
	ExecError     error
	AssertsResult []*AssertionResult
	Duration      time.Duration
	// DebugOutput holds the rendered manifests to show when debug is enabled for
	// this test job. It is intentionally left out of Stringify, so it does not end
	// up in the xml reports.
	DebugOutput string
}

// printDebugOutput prints the rendered manifests collected while running the job.
// The name is only printed when the rest of print doesn't, to avoid repeating it.
func (tjr TestJobResult) printDebugOutput(printer *printer.Printer, withName bool) {
	if tjr.DebugOutput == "" {
		return
	}
	if withName {
		printer.Println(printer.Faint("- %s", tjr.DisplayName), 1)
	}
	for line := range strings.SplitSeq(strings.TrimRight(tjr.DebugOutput, "\n"), "\n") {
		printer.Println(printer.Faint("%s", line), 2)
	}
	printer.Println("", 0)
}

// print the information to the console.
func (tjr TestJobResult) print(printer *printer.Printer, verbosity int) {
	if tjr.Passed {
		// nothing else is printed for a passing job, so label the debug output
		tjr.printDebugOutput(printer, true)
		return
	}

	if tjr.Skipped {
		msg := printer.Highlight("- ")
		msg += printer.WarningLabel("SKIPPED")
		msg += printer.Warning(" '%s'", tjr.DisplayName)
		printer.Println(msg, 1)
		tjr.printDebugOutput(printer, false)
		return
	}

	if tjr.ExecError != nil {
		printer.Println(printer.Highlight("- %s", tjr.DisplayName), 1)
		tjr.printDebugOutput(printer, false)
		printer.Println(printer.Highlight("Error: %s\n", tjr.ExecError.Error()), 2)
		return
	}

	printer.Println(printer.Danger("- %s\n", tjr.DisplayName), 1)
	tjr.printDebugOutput(printer, false)
	for _, assertResult := range tjr.AssertsResult {
		assertResult.print(printer, verbosity)
	}
}

// Stringify writing the object to a customized formatted string.
func (tjr TestJobResult) Stringify() string {
	var content strings.Builder

	if tjr.Skipped {
		fmt.Fprintf(&content, "SKIPPED '%s' \n", tjr.DisplayName)
	}

	if tjr.ExecError != nil {
		fmt.Fprintf(&content, "%s\n", tjr.ExecError.Error())
	}

	for _, assertResult := range tjr.AssertsResult {
		content.WriteString(assertResult.stringify())
	}

	return content.String()
}

// Stringify to xml attribute, replacing the object to a customized formatted string compatible with XML attributes.
func (tjr TestJobResult) StringifyToXmlAttribute() string {
	flattenString := strings.ReplaceAll(tjr.Stringify(), "\n", ",")
	return strings.ReplaceAll(flattenString, "\t", "")
}
