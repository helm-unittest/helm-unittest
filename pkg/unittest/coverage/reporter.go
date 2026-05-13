package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
)

// RenderConsole prints a human-friendly per-template coverage table to the
// provided printer. If p is nil, plain text is written to os.Stdout.
func RenderConsole(p *printer.Printer, cov Coverage) {
	w := io.Writer(os.Stdout)
	if p != nil {
		w = p.Writer
	}

	fmt.Fprintln(w)
	header := fmt.Sprintf("### Coverage [ %s ]", cov.ChartName)
	if p != nil {
		header = fmt.Sprintf("### Coverage [ %s ]", p.Highlight("%s", cov.ChartName))
	}
	fmt.Fprintln(w, header)
	fmt.Fprintln(w)

	if len(cov.Files) == 0 {
		fmt.Fprintln(w, "(no templates were instrumented)")
		return
	}

	rows := make([][]string, 0, len(cov.Files)+2)
	rows = append(rows, []string{"File", "Actions", "Branches", "Loops"})
	for _, f := range cov.Files {
		rows = append(rows, []string{
			f.Name,
			formatStat(p, f.Actions, f.ParseError),
			formatStat(p, f.Branches, f.ParseError),
			formatLoopStat(p, f.Loops, f.ParseError),
		})
	}
	rows = append(rows, []string{
		"ALL FILES",
		formatStat(p, cov.Totals.Actions, nil),
		formatStat(p, cov.Totals.Branches, nil),
		formatLoopStat(p, cov.Totals.Loops, nil),
	})

	printTable(w, rows)

	// Surface parse failures and uncovered lines as a follow-up block.
	var parseErrs []FileCoverage
	var uncovered []FileCoverage
	for _, f := range cov.Files {
		if f.ParseError != nil {
			parseErrs = append(parseErrs, f)
		} else if len(f.MissedLines) > 0 {
			uncovered = append(uncovered, f)
		}
	}
	if len(parseErrs) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Templates skipped (parse error):")
		for _, f := range parseErrs {
			fmt.Fprintf(w, "  - %s: %v\n", f.Name, f.ParseError)
		}
	}
	if len(uncovered) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Uncovered lines:")
		for _, f := range uncovered {
			fmt.Fprintf(w, "  - %s: %s\n", f.Name, joinInts(f.MissedLines))
		}
	}
	fmt.Fprintln(w)
}

func formatStat(p *printer.Printer, s CountStat, parseErr error) string {
	if parseErr != nil {
		return "parse-error"
	}
	if s.Total == 0 {
		return "-"
	}
	plain := fmt.Sprintf("%d/%d (%.1f%%)", s.Covered, s.Total, s.Pct())
	return colorizeByPct(p, plain, s.Pct())
}

// formatLoopStat augments the standard stat string with the total iteration
// count, so the table conveys "ran N times" alongside "X of Y loop bodies
// were entered". Iterations only show when at least one loop ran.
func formatLoopStat(p *printer.Printer, s CountStat, parseErr error) string {
	if parseErr != nil {
		return "parse-error"
	}
	if s.Total == 0 {
		return "-"
	}
	plain := fmt.Sprintf("%d/%d (%.1f%%)", s.Covered, s.Total, s.Pct())
	if s.Hits > 0 {
		plain = fmt.Sprintf("%s, %d iters", plain, s.Hits)
	}
	return colorizeByPct(p, plain, s.Pct())
}

func colorizeByPct(p *printer.Printer, text string, pct float64) string {
	if p == nil {
		return text
	}
	switch {
	case pct >= 80:
		return p.Success("%s", text)
	case pct >= 50:
		return p.Warning("%s", text)
	default:
		return p.Danger("%s", text)
	}
}

// printTable renders a left-aligned table. We avoid pulling a third-party
// table library to keep dependency surface minimal — the layout is simple
// enough to compute inline. Column widths are computed from visible-rune
// length so embedded ANSI color escapes do not throw off alignment.
func printTable(w io.Writer, rows [][]string) {
	if len(rows) == 0 {
		return
	}
	cols := len(rows[0])
	widths := make([]int, cols)
	for _, r := range rows {
		for i := 0; i < cols && i < len(r); i++ {
			if l := visibleLen(r[i]); l > widths[i] {
				widths[i] = l
			}
		}
	}
	for ri, r := range rows {
		var sb strings.Builder
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(r) {
				cell = r[i]
			}
			sb.WriteString(cell)
			sb.WriteString(strings.Repeat(" ", widths[i]-visibleLen(cell)))
			if i < cols-1 {
				sb.WriteString("  ")
			}
		}
		fmt.Fprintln(w, sb.String())
		if ri == 0 {
			// header separator
			var sep strings.Builder
			for i := 0; i < cols; i++ {
				sep.WriteString(strings.Repeat("-", widths[i]))
				if i < cols-1 {
					sep.WriteString("  ")
				}
			}
			fmt.Fprintln(w, sep.String())
		}
	}
}

// visibleLen returns the rune length of s with ANSI color escapes removed.
func visibleLen(s string) int {
	out := 0
	inEsc := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case inEsc:
			// skip
		default:
			out++
		}
	}
	return out
}

func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = fmt.Sprintf("%d", x)
	}
	return strings.Join(parts, ", ")
}

// Supported file-output formats for --coverage-format.
const (
	FormatJSON      = "json"
	FormatCobertura = "cobertura"
	FormatLCOV      = "lcov"
	FormatHTML      = "html"
)

// WriteReport dispatches to the writer for the requested format. An unknown
// format returns an error rather than silently picking a default — callers are
// expected to validate user input before calling.
func WriteReport(path, format string, cov Coverage) error {
	switch format {
	case "", FormatJSON:
		return WriteJSON(path, cov)
	case FormatCobertura:
		return WriteCobertura(path, cov)
	case FormatLCOV:
		return WriteLCOV(path, cov)
	case FormatHTML:
		return WriteHTML(path, cov)
	default:
		return fmt.Errorf("unsupported coverage format %q (want json, cobertura, lcov, or html)", format)
	}
}

// WriteJSON writes a stable JSON document to path describing per-file and
// total coverage. The schema is intended for CI consumption and is documented
// in DOCUMENT.md.
func WriteJSON(path string, cov Coverage) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(toJSONReport(cov))
}

type jsonStat struct {
	Covered int     `json:"covered"`
	Total   int     `json:"total"`
	Pct     float64 `json:"pct"`
	Hits    int64   `json:"hits"`
}

type jsonFile struct {
	Name        string   `json:"name"`
	ParseError  string   `json:"parseError,omitempty"`
	Actions     jsonStat `json:"actions"`
	Branches    jsonStat `json:"branches"`
	Loops       jsonStat `json:"loops"`
	MissedLines []int    `json:"missedLines,omitempty"`
}

type jsonReport struct {
	Chart  string   `json:"chart"`
	Files  []jsonFile `json:"files"`
	Totals struct {
		Actions  jsonStat `json:"actions"`
		Branches jsonStat `json:"branches"`
		Loops    jsonStat `json:"loops"`
	} `json:"totals"`
}

func toJSONReport(cov Coverage) jsonReport {
	r := jsonReport{Chart: cov.ChartName}
	for _, f := range cov.Files {
		entry := jsonFile{
			Name:        f.Name,
			Actions:     toJSONStat(f.Actions),
			Branches:    toJSONStat(f.Branches),
			Loops:       toJSONStat(f.Loops),
			MissedLines: f.MissedLines,
		}
		if f.ParseError != nil {
			entry.ParseError = f.ParseError.Error()
		}
		r.Files = append(r.Files, entry)
	}
	r.Totals.Actions = toJSONStat(cov.Totals.Actions)
	r.Totals.Branches = toJSONStat(cov.Totals.Branches)
	r.Totals.Loops = toJSONStat(cov.Totals.Loops)
	return r
}

func toJSONStat(s CountStat) jsonStat {
	pct := s.Pct()
	if pct < 0 {
		pct = 0
	}
	return jsonStat{Covered: s.Covered, Total: s.Total, Pct: pct, Hits: s.Hits}
}
