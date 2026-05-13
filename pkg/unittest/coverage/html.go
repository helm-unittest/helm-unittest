package coverage

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"os"
	"strings"
)

// WriteHTML writes a self-contained HTML coverage report at path. The output
// has no external dependencies (no JS, no CDN-hosted CSS) so it works as a
// CI artifact or as a file opened directly in a browser. Each template's
// source is embedded inline with per-line colour coding sourced from the
// per-probe hit counts; clicking a row in the file list jumps to that file's
// source block.
func WriteHTML(path string, cov Coverage) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return renderHTML(f, cov)
}

func renderHTML(w io.Writer, cov Coverage) error {
	var b strings.Builder
	writeHTMLHead(&b, cov)
	writeHTMLSummary(&b, cov)
	writeHTMLFileList(&b, cov)
	for _, file := range cov.Files {
		writeHTMLFileView(&b, file)
	}
	writeHTMLFoot(&b)
	_, err := io.WriteString(w, b.String())
	return err
}

const htmlStyles = `
  :root { color-scheme: light dark; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
         margin: 24px; color: #1f2328; background: #ffffff; }
  h1 { margin: 0 0 4px 0; font-size: 22px; }
  .subtitle { color: #57606a; margin-bottom: 24px; font-size: 13px; }
  .summary { display: flex; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
  .stat { background: #f6f8fa; border: 1px solid #d0d7de; border-radius: 6px;
          padding: 10px 16px; min-width: 140px; }
  .stat .label { font-size: 11px; color: #57606a; text-transform: uppercase; letter-spacing: .04em; }
  .stat .value { font-size: 22px; font-weight: 600; margin-top: 2px; }
  .stat .sub { font-size: 12px; color: #57606a; margin-top: 2px; }
  table.files { width: 100%; border-collapse: collapse; margin-bottom: 32px; font-size: 13px; }
  .files th, .files td { text-align: left; padding: 6px 12px; border-bottom: 1px solid #d0d7de; }
  .files th { font-weight: 600; background: #f6f8fa; }
  .files th.num, .files td.num { text-align: right; font-variant-numeric: tabular-nums; }
  .files a { color: #0969da; text-decoration: none; }
  .files a:hover { text-decoration: underline; }
  .pct-high { color: #1a7f37; }
  .pct-mid  { color: #9a6700; }
  .pct-low  { color: #cf222e; }
  .pct-na   { color: #6e7781; }
  details.file { margin: 10px 0; border: 1px solid #d0d7de; border-radius: 6px; overflow: hidden; }
  details.file summary { background: #f6f8fa; padding: 8px 12px; cursor: pointer; font-weight: 600;
                         display: flex; gap: 16px; align-items: baseline; }
  details.file summary .badges { margin-left: auto; font-weight: 400; font-size: 12px; color: #57606a; }
  table.source { width: 100%; border-collapse: collapse;
                 font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12px; }
  .source td { padding: 0 8px; vertical-align: top; }
  .source .lineno { color: #6e7781; text-align: right; width: 1%; user-select: none;
                    background: #f6f8fa; border-right: 1px solid #d0d7de; }
  .source .hits   { color: #6e7781; text-align: right; width: 1%; font-variant-numeric: tabular-nums;
                    background: #f6f8fa; border-right: 1px solid #d0d7de; }
  .source .code   { white-space: pre; }
  .source tr.covered .code { background: #ddf4e4; }
  .source tr.missed  .code { background: #ffebe9; }
  .source tr.partial .code { background: #fff8c5; }
  .source tr.neutral .code { background: transparent; }
  .parse-error { padding: 10px 14px; background: #ffebe9; color: #cf222e; }
  @media (prefers-color-scheme: dark) {
    body { color: #c9d1d9; background: #0d1117; }
    .stat, .files th, details.file summary,
    .source .lineno, .source .hits { background: #161b22; border-color: #30363d; }
    .files th, .files td { border-bottom-color: #30363d; }
    details.file { border-color: #30363d; }
    .source tr.covered .code { background: #1b4721; }
    .source tr.missed  .code { background: #5a1d1a; }
    .source tr.partial .code { background: #5c4a00; }
    .parse-error { background: #5a1d1a; color: #ffa198; }
  }
`

func writeHTMLHead(b *strings.Builder, cov Coverage) {
	fmt.Fprintf(b, `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>helm-unittest coverage: %s</title>
<style>%s</style>
</head>
<body>
<h1>Coverage: %s</h1>
<div class="subtitle">helm-unittest — generated report</div>
`, html.EscapeString(cov.ChartName), htmlStyles, html.EscapeString(cov.ChartName))
}

func writeHTMLFoot(b *strings.Builder) {
	b.WriteString("</body>\n</html>\n")
}

func writeHTMLSummary(b *strings.Builder, cov Coverage) {
	b.WriteString(`<section class="summary">`)
	statCard(b, "Actions", cov.Totals.Actions, false)
	statCard(b, "Branches", cov.Totals.Branches, false)
	statCard(b, "Loops", cov.Totals.Loops, true)
	b.WriteString(`</section>`)
}

func statCard(b *strings.Builder, label string, s CountStat, showIters bool) {
	pct := s.Pct()
	pctClass := pctClass(pct)
	pctText := "N/A"
	subText := ""
	if s.Total > 0 {
		pctText = fmt.Sprintf("%.1f%%", pct)
		subText = fmt.Sprintf("%d / %d covered", s.Covered, s.Total)
	}
	fmt.Fprintf(b,
		`<div class="stat"><div class="label">%s</div><div class="value %s">%s</div><div class="sub">%s`,
		html.EscapeString(label), pctClass, pctText, html.EscapeString(subText))
	if showIters && s.Hits > 0 {
		fmt.Fprintf(b, ` &middot; %d iters`, s.Hits)
	}
	b.WriteString(`</div></div>`)
}

func writeHTMLFileList(b *strings.Builder, cov Coverage) {
	if len(cov.Files) == 0 {
		b.WriteString(`<p>(no templates were instrumented)</p>`)
		return
	}
	b.WriteString(`<table class="files"><thead><tr><th>File</th><th class="num">Actions</th><th class="num">Branches</th><th class="num">Loops</th><th class="num">Iters</th></tr></thead><tbody>`)
	for _, f := range cov.Files {
		fmt.Fprintf(b,
			`<tr><td><a href="#%s">%s</a></td><td class="num">%s</td><td class="num">%s</td><td class="num">%s</td><td class="num">%s</td></tr>`,
			htmlID(f.Name),
			html.EscapeString(f.Name),
			statCell(f.Actions, f.ParseError),
			statCell(f.Branches, f.ParseError),
			statCell(f.Loops, f.ParseError),
			itersCell(f.Loops, f.ParseError),
		)
	}
	b.WriteString(`</tbody></table>`)
}

func statCell(s CountStat, parseErr error) string {
	if parseErr != nil {
		return `<span class="pct-na">parse-error</span>`
	}
	if s.Total == 0 {
		return `<span class="pct-na">&mdash;</span>`
	}
	return fmt.Sprintf(`<span class="%s">%d/%d (%.1f%%)</span>`,
		pctClass(s.Pct()), s.Covered, s.Total, s.Pct())
}

func itersCell(loops CountStat, parseErr error) string {
	if parseErr != nil || loops.Total == 0 || loops.Hits == 0 {
		return `<span class="pct-na">&mdash;</span>`
	}
	return fmt.Sprintf(`%d`, loops.Hits)
}

func writeHTMLFileView(b *strings.Builder, f FileCoverage) {
	fmt.Fprintf(b, `<details class="file" id="%s"><summary>%s<span class="badges">%s</span></summary>`,
		htmlID(f.Name), html.EscapeString(f.Name), summaryBadges(f))

	if f.ParseError != nil {
		fmt.Fprintf(b, `<div class="parse-error">parse error: %s</div></details>`,
			html.EscapeString(f.ParseError.Error()))
		return
	}

	lineMap := map[int]LineCoverage{}
	for _, ln := range f.Lines {
		lineMap[ln.Line] = ln
	}

	b.WriteString(`<table class="source"><tbody>`)
	lines := bytes.Split(f.Source, []byte("\n"))
	for i, raw := range lines {
		lineNum := i + 1
		class := "neutral"
		hits := "."
		if lc, ok := lineMap[lineNum]; ok {
			switch {
			case lc.ProbesCovered == 0:
				class = "missed"
			case lc.ProbesCovered == lc.ProbesTotal:
				class = "covered"
			default:
				class = "partial"
			}
			hits = fmt.Sprintf("%d", lc.Hits)
		}
		fmt.Fprintf(b,
			`<tr class="%s"><td class="lineno">%d</td><td class="hits">%s</td><td class="code">%s</td></tr>`,
			class, lineNum, hits, html.EscapeString(string(raw)))
	}
	b.WriteString(`</tbody></table></details>`)
}

func summaryBadges(f FileCoverage) string {
	if f.ParseError != nil {
		return `<span class="pct-low">parse error</span>`
	}
	parts := []string{}
	if f.Actions.Total > 0 {
		parts = append(parts, fmt.Sprintf(`A %d/%d`, f.Actions.Covered, f.Actions.Total))
	}
	if f.Branches.Total > 0 {
		parts = append(parts, fmt.Sprintf(`B %d/%d`, f.Branches.Covered, f.Branches.Total))
	}
	if f.Loops.Total > 0 {
		parts = append(parts, fmt.Sprintf(`L %d/%d`, f.Loops.Covered, f.Loops.Total))
	}
	return html.EscapeString(strings.Join(parts, " · "))
}

func pctClass(pct float64) string {
	switch {
	case pct < 0:
		return "pct-na"
	case pct >= 80:
		return "pct-high"
	case pct >= 50:
		return "pct-mid"
	default:
		return "pct-low"
	}
}

// htmlID maps a template path to an HTML anchor id. We strip characters that
// confuse `#fragment` URLs (dots, slashes) while keeping the result readable.
func htmlID(name string) string {
	var b strings.Builder
	b.WriteString("f-")
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}
