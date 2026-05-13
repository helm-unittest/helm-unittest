package coverage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleHTMLCoverage() Coverage {
	cov := Coverage{ChartName: "demo"}
	cov.Files = []FileCoverage{
		{
			Name:     "demo/templates/cm.yaml",
			Rendered: true,
			Actions:  CountStat{Covered: 1, Total: 2, Hits: 5},
			Branches: CountStat{Covered: 1, Total: 2, Hits: 3},
			Loops:    CountStat{Covered: 1, Total: 1, Hits: 4},
			Source:   []byte("apiVersion: v1\nkind: ConfigMap\nname: {{ .Values.name }}\n"),
			Lines: []LineCoverage{
				{Line: 3, Hits: 5, ProbesCovered: 1, ProbesTotal: 1},
			},
		},
		{
			Name:     "demo/templates/dead.yaml",
			Rendered: false,
			Source:   []byte("apiVersion: v1\nkind: ConfigMap\n"),
		},
		{
			Name:       "demo/templates/broken.yaml",
			ParseError: errors.New("synthetic parse failure"),
			Source:     []byte("{{ if .Values.x"),
		},
	}
	cov.Totals.Actions = CountStat{Covered: 1, Total: 2, Hits: 5}
	cov.Totals.Branches = CountStat{Covered: 1, Total: 2, Hits: 3}
	cov.Totals.Loops = CountStat{Covered: 1, Total: 1, Hits: 4}
	return cov
}

func TestWriteHTML_StructureAndContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cov.html")
	require.NoError(t, WriteHTML(path, sampleHTMLCoverage()))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	contents := string(data)

	// Doctype + correct chart name in <title> and <h1>.
	assert.True(t, strings.HasPrefix(contents, "<!doctype html>"), "missing doctype")
	assert.Contains(t, contents, "<title>helm-unittest coverage: demo</title>")
	assert.Contains(t, contents, "Coverage: demo")

	// Summary cards for each dimension and the iterations badge.
	assert.Contains(t, contents, ">Actions<")
	assert.Contains(t, contents, ">Branches<")
	assert.Contains(t, contents, ">Loops<")
	assert.Contains(t, contents, "4 iters", "iteration count should appear on the loops stat card")

	// File-list rows.
	assert.Contains(t, contents, `<a href="#f-demo-templates-cm-yaml">demo/templates/cm.yaml</a>`)
	assert.Contains(t, contents, `<a href="#f-demo-templates-broken-yaml">demo/templates/broken.yaml</a>`)

	// Source lines: line 3 has a covered probe so should carry the covered class.
	assert.Contains(t, contents, `<tr class="covered"><td class="lineno">3</td>`)
	// Lines 1 and 2 have no probes — neutral.
	assert.Contains(t, contents, `<tr class="neutral"><td class="lineno">1</td>`)

	// Source content must be HTML-escaped so {{ and < survive intact.
	assert.Contains(t, contents, "{{ .Values.name }}")
	// Parse-error file gets a parse-error block, not a source table.
	assert.Contains(t, contents, "synthetic parse failure")
	// And no source table for the broken file.
	assert.NotContains(t, contents, `<a href="#f-demo-templates-broken-yaml"><table`)

	// Rendered/unused badges must appear on the right files.
	assert.Contains(t, contents, `<span class="badge badge-used">used</span>`,
		"cm.yaml should show a 'used' badge")
	assert.Contains(t, contents, `<span class="badge badge-unused">unused</span>`,
		"dead.yaml should show an 'unused' badge")
	// Each badge should appear exactly twice (once in the file list row,
	// once in the per-file <details> summary). Match the full element so the
	// `.badge-used` / `.badge-unused` rules in the embedded stylesheet don't
	// throw the count off.
	usedCount := strings.Count(contents, `<span class="badge badge-used">used</span>`)
	unusedCount := strings.Count(contents, `<span class="badge badge-unused">unused</span>`)
	assert.Equal(t, 2, usedCount, "cm.yaml shows used badge twice (file list + summary)")
	assert.Equal(t, 2, unusedCount, "dead.yaml shows unused badge twice")
}

func TestWriteReport_HTMLDispatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.html")
	require.NoError(t, WriteReport(path, FormatHTML, sampleHTMLCoverage()))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "<!doctype html>"),
		"WriteReport(..., \"html\", ...) should dispatch to WriteHTML")
}

func TestHTMLID_StableAndSafe(t *testing.T) {
	cases := map[string]string{
		"demo/templates/cm.yaml":         "f-demo-templates-cm-yaml",
		"chart-with-dashes/templates/x":  "f-chart-with-dashes-templates-x",
		"sub/_helpers.tpl":               "f-sub--helpers-tpl",
	}
	for in, want := range cases {
		assert.Equal(t, want, htmlID(in), in)
	}
}
