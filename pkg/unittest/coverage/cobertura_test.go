package coverage

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleCoverage() Coverage {
	cov := Coverage{ChartName: "demo"}
	cov.Files = []FileCoverage{
		{
			Name:    "demo/templates/cm.yaml",
			Actions: CountStat{Covered: 2, Total: 3},
			Branches: CountStat{Covered: 1, Total: 2},
			Loops:   CountStat{Covered: 0, Total: 1},
			Lines: []LineCoverage{
				{Line: 4, Hits: 2},
				{Line: 6, Hits: 0, Branches: []BranchCoverage{{Label: "if", Hits: 0}, {Label: "else", Hits: 2}}},
				{Line: 9, Hits: 0, Branches: []BranchCoverage{{Label: "range-body", Hits: 0}}},
			},
		},
		{
			Name:       "demo/templates/broken.yaml",
			ParseError: errors.New("synthetic"),
		},
	}
	cov.Totals.Actions = CountStat{Covered: 2, Total: 3}
	cov.Totals.Branches = CountStat{Covered: 1, Total: 2}
	cov.Totals.Loops = CountStat{Covered: 0, Total: 1}
	return cov
}

func TestWriteCobertura_StructureAndCounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cobertura.xml")
	require.NoError(t, WriteCobertura(path, sampleCoverage()))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	contents := string(data)

	// Header + DOCTYPE present so Cobertura-aware tools accept the file.
	assert.True(t, strings.HasPrefix(contents, `<?xml`), "missing XML header")
	assert.Contains(t, contents, "coverage-04.dtd")

	var doc coberturaCoverage
	require.NoError(t, xml.Unmarshal(data, &doc))

	// Top-level totals come from actions (lines) and branches+loops combined.
	assert.Equal(t, 2, doc.LinesCovered)
	assert.Equal(t, 3, doc.LinesValid)
	assert.Equal(t, 1, doc.BranchesCovered)
	assert.Equal(t, 3, doc.BranchesValid)

	require.Len(t, doc.Packages.Packages, 1)
	pkg := doc.Packages.Packages[0]
	assert.Equal(t, "demo", pkg.Name)

	// Parse-error files must not appear as classes in the package.
	require.Len(t, pkg.Classes.Classes, 1)
	cls := pkg.Classes.Classes[0]
	assert.Equal(t, "demo/templates/cm.yaml", cls.Filename)
	require.Len(t, cls.Lines.Lines, 3)

	// Branch line should be marked and carry condition-coverage attribute.
	branchLine := cls.Lines.Lines[1]
	assert.Equal(t, 6, branchLine.Number)
	assert.Equal(t, "true", branchLine.Branch)
	assert.Contains(t, branchLine.ConditionCoverage, "1/2")
}

func TestClassNameFromPath(t *testing.T) {
	cases := map[string]string{
		"demo/templates/cm.yaml":              "demo.templates.cm_yaml",
		"demo/templates/sub/dir/cm.yaml":      "demo.templates.sub.dir.cm_yaml",
		"cm.yaml":                             "cm_yaml",
	}
	for in, want := range cases {
		assert.Equal(t, want, classNameFromPath(in), in)
	}
}
