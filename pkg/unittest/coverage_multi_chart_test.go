package unittest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: a multi-chart run must write one merged report per format,
// since each chart's report file would otherwise collide on the same path.
func TestV4RunnerCoverageMultiChartWritesOneMergedReport(t *testing.T) {
	buffer := new(bytes.Buffer)
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "cov.xml")

	runner := TestRunner{
		Printer:        printer.NewPrinter(buffer, nil),
		Coverage:       true,
		WithSubChart:   true,
		TestFiles:      []string{"tests/*_test.yaml"},
		CoverageFile:   reportPath,
		CoverageFormat: "cobertura",
	}

	passed := runner.RunV4([]string{
		"../../test/data/v3/basic",
		"../../test/data/v3/coverage-fromjson",
	})
	require.True(t, passed, buffer.String())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "expected exactly one report file, got %v", entries)
	assert.Equal(t, "cov.xml", entries[0].Name())

	data, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	contents := string(data)
	assert.Contains(t, contents, "basic/templates", "first chart missing from merged report")
	assert.Contains(t, contents, "coverage-fromjson/templates", "second chart missing from merged report")
	assert.True(t, strings.HasPrefix(contents, "<?xml"))
}
