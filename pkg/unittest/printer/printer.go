package printer

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// NewPrinter create a Printer with Writer to print and colored config
func NewPrinter(writer io.Writer, colored *bool) *Printer {
	p := &Printer{
		Writer:  writer,
		Colored: colored,
	}

	p.colors.Success = color.New(color.FgGreen)
	p.setupColor(p.colors.Success)
	p.colors.SuccessBg = color.New(color.BgGreen, color.FgBlack)
	p.setupColor(p.colors.SuccessBg)
	p.colors.Danger = color.New(color.FgRed)
	p.setupColor(p.colors.Danger)
	p.colors.DangerBg = color.New(color.BgRed, color.FgWhite)
	p.setupColor(p.colors.DangerBg)
	p.colors.Warning = color.New(color.FgYellow)
	p.setupColor(p.colors.Warning)
	p.colors.WarningBg = color.New(color.BgYellow, color.FgBlack)
	p.setupColor(p.colors.WarningBg)
	p.colors.Highlight = color.New(color.Bold)
	p.setupColor(p.colors.Highlight)
	p.colors.Faint = color.New(color.Faint)
	p.setupColor(p.colors.Faint)

	return p
}

// Printer simple printing implement
type Printer struct {
	Writer  io.Writer
	Colored *bool
	colors  struct {
		Success   *color.Color
		SuccessBg *color.Color
		Warning   *color.Color
		WarningBg *color.Color
		Danger    *color.Color
		DangerBg  *color.Color
		Highlight *color.Color
		Faint     *color.Color
	}
}

func (p *Printer) setupColor(color *color.Color) {
	if p.Colored != nil {
		if *p.Colored {
			color.EnableColor()
		} else {
			color.DisableColor()
		}
	}
}

// Println print a line with indent level.
func (p *Printer) Println(content string, indentLevel int) {
	var indent string
	for range indentLevel {
		indent += "\t"
	}
	_, _ = fmt.Fprintln(p.Writer, indent+content)
}

// Success print success.
func (p *Printer) Success(format string, a ...any) string {
	return p.colors.Success.Sprintf(format, a...)
}

// SuccessLabel print success as label.
func (p *Printer) SuccessLabel(format string, a ...any) string {
	return p.colors.SuccessBg.Sprintf(format, a...)
}

// Danger print danger.
func (p *Printer) Danger(format string, a ...any) string {
	return p.colors.Danger.Sprintf(format, a...)
}

// DangerLabel print danger as label.
func (p *Printer) DangerLabel(format string, a ...any) string {
	return p.colors.DangerBg.Sprintf(format, a...)
}

// Warning print warning.
func (p *Printer) Warning(format string, a ...any) string {
	return p.colors.Warning.Sprintf(format, a...)
}

// WarningLabel print warning as label.
func (p *Printer) WarningLabel(format string, a ...any) string {
	return p.colors.WarningBg.Sprintf(format, a...)
}

// Highlight print highlight.
func (p *Printer) Highlight(format string, a ...any) string {
	return p.colors.Highlight.Sprintf(format, a...)
}

// Faint print faint.
func (p *Printer) Faint(format string, a ...any) string {
	return p.colors.Faint.Sprintf(format, a...)
}
