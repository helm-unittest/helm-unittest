package unittest

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// loggable simple printing interface
type loggable interface {
	println(content string, indentLevel int)
	success(format string, a ...interface{}) string
	successLabel(format string, a ...interface{}) string
	danger(format string, a ...interface{}) string
	dangerLabel(format string, a ...interface{}) string
	warning(format string, a ...interface{}) string
	warningLabel(format string, a ...interface{}) string
	highlight(format string, a ...interface{}) string
	faint(format string, a ...interface{}) string
}

// Printer simple printing implement
type Printer struct {
	Writer       io.Writer
	Colored      bool
	ColredForced bool
}

func (p *Printer) println(content string, indentLevel int) {
	var indent string
	for i := 0; i < indentLevel; i++ {
		indent += "\t"
	}
	fmt.Fprintln(p.Writer, indent+content)
}

func (p Printer) success(format string, a ...interface{}) string {
	return color.GreenString(format, a...)
}

var greenBgBlackFg = color.New(color.BgGreen, color.FgBlack)

func (p Printer) successLabel(format string, a ...interface{}) string {
	if p.Colored {
		return greenBgBlackFg.Sprintf(format, a...)
	}
	return "[" + fmt.Sprintf(format, a...) + "]"
}

func (p Printer) danger(format string, a ...interface{}) string {
	return color.RedString(format, a...)
}

var redBgWhiteFg = color.New(color.BgRed, color.FgWhite)

func (p Printer) dangerLabel(format string, a ...interface{}) string {
	if p.Colored {
		return redBgWhiteFg.Sprintf(format, a...)
	}
	return "[" + fmt.Sprintf(format, a...) + "]"
}

func (p Printer) warning(format string, a ...interface{}) string {
	return color.YellowString(format, a...)
}

var yellowBgBlackFg = color.New(color.BgYellow, color.FgBlack)

func (p Printer) warningLabel(format string, a ...interface{}) string {
	if p.Colored {
		return yellowBgBlackFg.Sprintf(format, a...)
	}
	return "[" + fmt.Sprintf(format, a...) + "]"
}

var bold = color.New(color.Bold)

func (p Printer) highlight(format string, a ...interface{}) string {
	return bold.Sprintf(format, a...)
}

var faint = color.New(color.Faint)

func (p Printer) faint(format string, a ...interface{}) string {
	return faint.Sprintf(format, a...)
}
