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
}

// print the information to the console.
func (tjr TestJobResult) print(printer *printer.Printer, verbosity int) {
	if tjr.Passed {
		return
	}

	if tjr.Skipped {
		msg := printer.Highlight("- ")
		msg += printer.WarningLabel("SKIPPED")
		msg += printer.Warning(" '%s'", tjr.DisplayName)
		printer.Println(msg, 1)
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

// Stringify to xml attribute, replacing the the object to a customized formatted string.
func (tjr TestJobResult) StringifyToXmlAttribute() string {
	flattenString := strings.ReplaceAll(tjr.Stringify(), "\n", ",")
	return strings.ReplaceAll(flattenString, "\t", "")
}
