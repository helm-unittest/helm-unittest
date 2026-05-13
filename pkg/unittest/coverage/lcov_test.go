package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteLCOV_RecordsAndCounters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lcov.info")
	require.NoError(t, WriteLCOV(path, sampleCoverage()))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	contents := string(data)

	// Two records: one real, one empty for the parse-error file.
	assert.Equal(t, 2, strings.Count(contents, "end_of_record"))
	assert.Contains(t, contents, "SF:demo/templates/cm.yaml")
	assert.Contains(t, contents, "SF:demo/templates/broken.yaml")

	// Per-line DA lines must be present for every Lines entry on the real file.
	assert.Contains(t, contents, "DA:4,2")
	assert.Contains(t, contents, "DA:6,0")
	assert.Contains(t, contents, "DA:9,0")

	// Branch records: line 6 has two branches (if covered=0, else covered=2),
	// line 9 has one (range body, uncovered).
	assert.Contains(t, contents, "BRDA:6,6,0,-")    // if: uncovered
	assert.Contains(t, contents, "BRDA:6,6,1,2")    // else: hit 2x
	assert.Contains(t, contents, "BRDA:9,9,0,-")    // range body: uncovered

	// Summary lines for the real record.
	assert.Contains(t, contents, "LF:3")
	assert.Contains(t, contents, "LH:1") // only line 4 had hits
	assert.Contains(t, contents, "BRF:3")
	assert.Contains(t, contents, "BRH:1") // only the else branch ran
}

func TestWriteReport_UnknownFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out")
	err := WriteReport(path, "nope", sampleCoverage())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported coverage format")
}

func TestWriteReport_DefaultIsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	require.NoError(t, WriteReport(path, "", sampleCoverage()))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(data)), "{"), "default writer should emit JSON")
}
