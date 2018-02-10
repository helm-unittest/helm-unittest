package helmtest

import "fmt"

type AssertionResult struct {
	Index      int
	FailInfo   []string
	Passed     bool
	AssertType string
	CustomInfo string
}

func (ar AssertionResult) print(printer loggable, verbosity int) {
	if ar.Passed {
		return
	}
	var title string
	if ar.CustomInfo != "" {
		title = ar.CustomInfo
	} else {
		title = fmt.Sprintf("- asserts[%d] `%s` fail", ar.Index, ar.AssertType)
	}
	printer.println(printer.danger(title+"\n"), 2)
	for _, infoLine := range ar.FailInfo {
		printer.println(infoLine, 3)
	}
	printer.println("", 0)
}
