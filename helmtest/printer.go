package helmtest

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type loggable interface {
	println(content string, indentLevel int)
	success(content string) string
	successBackground(content string) string
	danger(content string) string
	dangerBackground(content string) string
	warning(content string) string
	warningBackground(content string) string
	highlight(content string) string
}

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

func (p Printer) success(content string) string {
	return color.GreenString(content)
}

var greenBgBlackFg = color.New(color.BgGreen, color.FgBlack)

func (p Printer) successBackground(content string) string {
	if p.Colored {
		return greenBgBlackFg.Sprint(content)
	}
	return "[" + content + "]"
}

func (p Printer) danger(content string) string {
	return color.RedString(content)
}

var redBgWhiteFg = color.New(color.BgRed, color.FgWhite)

func (p Printer) dangerBackground(content string) string {
	if p.Colored {
		return redBgWhiteFg.Sprint(content)
	}
	return "[" + content + "]"
}

func (p Printer) warning(content string) string {
	return color.YellowString(content)
}

var yellowBgBlackFg = color.New(color.BgYellow, color.FgBlack)

func (p Printer) warningBackground(content string) string {
	if p.Colored {
		return yellowBgBlackFg.Sprint(content)
	}
	return "[" + content + "]"
}

func (p Printer) highlight(content string) string {
	return color.WhiteString(content)
}
