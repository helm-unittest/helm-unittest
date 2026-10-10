package coverage_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest/coverage"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

func newPlainPrinter(buf *bytes.Buffer) *printer.Printer {
	colored := false
	return printer.NewPrinter(buf, &colored)
}

func TestRenderConsole_NoFiles(t *testing.T) {
	var buf bytes.Buffer
	RenderConsole(newPlainPrinter(&buf), Coverage{ChartName: "empty"})

	out := buf.String()
	assert.Contains(t, out, "### Coverage [ empty ]")
	assert.Contains(t, out, "(no templates were instrumented)")
	assert.NotContains(t, out, "ALL FILES")
}

func TestRenderConsole_Table(t *testing.T) {
	cov := Coverage{
		ChartName: "demo",
		Files: []FileCoverage{
			{
				Name:     "demo/templates/cm.yaml",
				Rendered: true,
				Actions:  CountStat{Covered: 2, Total: 3},
				Branches: CountStat{Covered: 1, Total: 2},
				Loops:    CountStat{Covered: 1, Total: 1, Hits: 4},
			},
			{
				Name:     "demo/templates/dead.yaml",
				Rendered: false,
				Actions:  CountStat{Covered: 0, Total: 1},
			},
		},
	}
	cov.Totals.Actions = CountStat{Covered: 2, Total: 4}
	cov.Totals.Branches = CountStat{Covered: 1, Total: 2}
	cov.Totals.Loops = CountStat{Covered: 1, Total: 1, Hits: 4}
	var buf bytes.Buffer
	RenderConsole(newPlainPrinter(&buf), cov)

	out := buf.String()
	for _, want := range []string{
		"File", "Actions", "Branches", "Loops", "Used",
		"demo/templates/cm.yaml", "2/3 (66.7%)", "1/2 (50.0%)", "1/1 (100.0%), 4 iters",
		"demo/templates/dead.yaml", "0/1 (0.0%)",
		"ALL FILES", "2/4 (50.0%)", "1/2",
	} {
		assert.Contains(t, out, want)
	}
	assert.Contains(t, out, "yes")
	assert.Contains(t, out, "no")
	assert.Contains(t, out, "---")
	assert.NotContains(t, out, "Templates skipped")
	assert.NotContains(t, out, "Uncovered lines")
}

func TestRenderConsole_ParseErrorAndUncovered(t *testing.T) {
	cov := Coverage{
		ChartName: "demo",
		Files: []FileCoverage{
			{
				Name:        "demo/templates/cm.yaml",
				Rendered:    true,
				Actions:     CountStat{Covered: 1, Total: 3},
				MissedLines: []int{4, 6, 9},
			},
			{
				Name:       "demo/templates/broken.yaml",
				ParseError: errors.New("synthetic"),
			},
		},
	}
	var buf bytes.Buffer
	RenderConsole(newPlainPrinter(&buf), cov)

	out := buf.String()
	assert.Contains(t, out, "parse-error")
	assert.Contains(t, out, "Templates skipped (parse error):")
	assert.Contains(t, out, "  - demo/templates/broken.yaml: synthetic")
	assert.Contains(t, out, "Uncovered lines:")
	assert.Contains(t, out, "  - demo/templates/cm.yaml: 4, 6, 9")
	assert.Contains(t, out, "1/2") // only the rendered file counts as used
}

func TestRenderConsole_NilPrinterWritesToStdout(t *testing.T) {
	// Must not panic and must fall back to os.Stdout.
	assert.NotPanics(t, func() {
		RenderConsole(nil, Coverage{ChartName: "demo"})
	})
}

func TestRenderConsole_ColoredOutput(t *testing.T) {
	colored := true
	var buf bytes.Buffer
	p := printer.NewPrinter(&buf, &colored)
	cov := Coverage{
		ChartName: "demo",
		Files: []FileCoverage{{
			Name:     "demo/templates/cm.yaml",
			Rendered: true,
			Actions:  CountStat{Covered: 3, Total: 3},
		}},
	}
	cov.Totals.Actions = CountStat{Covered: 3, Total: 3}
	RenderConsole(p, cov)

	out := buf.String()
	assert.Contains(t, out, "\x1b[", "expected ANSI escapes when colored")
	assert.Contains(t, out, "3/3 (100.0%)")
}
