package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormats(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		err  bool
	}{
		{in: "", want: []string{FormatJSON}},
		{in: "   ", want: []string{FormatJSON}},
		{in: "json", want: []string{FormatJSON}},
		{in: "cobertura,lcov,html", want: []string{FormatCobertura, FormatLCOV, FormatHTML}},
		{in: "cobertura , lcov , html", want: []string{FormatCobertura, FormatLCOV, FormatHTML}, // whitespace tolerated
		},
		{in: "html,html,html", want: []string{FormatHTML}}, // duplicates collapsed
		{in: "json,nope", err: true},
		{in: ",,,", err: true}, // nothing usable
	}
	for _, c := range cases {
		got, err := ParseFormats(c.in)
		if c.err {
			require.Error(t, err, "input %q should fail", c.in)
			continue
		}
		require.NoError(t, err, "input %q", c.in)
		assert.Equal(t, c.want, got, "input %q", c.in)
	}
}

func TestResolveOutputPaths_SingleFormatVerbatim(t *testing.T) {
	// One format → user's path is used exactly as given. This preserves
	// the long-standing single-format behaviour.
	got := ResolveOutputPaths("coverage.xml", []string{FormatCobertura})
	require.Len(t, got, 1)
	assert.Equal(t, "coverage.xml", got[0].Path)
	assert.Equal(t, FormatCobertura, got[0].Format)
}

func TestResolveOutputPaths_MultiFormatStem(t *testing.T) {
	// Multiple formats → path is a stem; each format gets its own extension.
	got := ResolveOutputPaths("./reports/cov", []string{FormatCobertura, FormatLCOV, FormatHTML, FormatJSON})
	want := map[string]string{
		"./reports/cov.xml":   FormatCobertura,
		"./reports/cov.info":  FormatLCOV,
		"./reports/cov.html":  FormatHTML,
		"./reports/cov.json":  FormatJSON,
	}
	for _, target := range got {
		assert.Equal(t, want[target.Path], target.Format, target.Path)
	}
	assert.Len(t, got, 4)
}

func TestResolveOutputPaths_MultiFormatStripsKnownExtension(t *testing.T) {
	// If the user accidentally passes a path with a known extension, we
	// strip it so we don't end up writing `coverage.xml.xml`.
	got := ResolveOutputPaths("coverage.xml", []string{FormatCobertura, FormatHTML})
	paths := []string{got[0].Path, got[1].Path}
	assert.Contains(t, paths, "coverage.xml")
	assert.Contains(t, paths, "coverage.html")
}

func TestResolveOutputPaths_MultiFormatUnknownExtensionKept(t *testing.T) {
	// Unknown extensions are left alone (treated as part of the stem).
	got := ResolveOutputPaths("custom.cov", []string{FormatCobertura, FormatLCOV})
	paths := []string{got[0].Path, got[1].Path}
	assert.Contains(t, paths, "custom.cov.xml")
	assert.Contains(t, paths, "custom.cov.info")
}

func TestFormatExt(t *testing.T) {
	assert.Equal(t, ".json", FormatExt(FormatJSON))
	assert.Equal(t, ".xml", FormatExt(FormatCobertura))
	assert.Equal(t, ".info", FormatExt(FormatLCOV))
	assert.Equal(t, ".html", FormatExt(FormatHTML))
	assert.Equal(t, "", FormatExt("bogus"))
}

// TestWriteReport_AcceptsAllResolvedFormats verifies the end-to-end multi-
// format flow: ResolveOutputPaths produces targets, WriteReport dispatches
// each one, and the resulting files have the right magic.
func TestWriteReport_AcceptsAllResolvedFormats(t *testing.T) {
	dir := t.TempDir()
	stem := filepath.Join(dir, "cov")

	formats, err := ParseFormats("cobertura,lcov,html,json")
	require.NoError(t, err)

	for _, target := range ResolveOutputPaths(stem, formats) {
		require.NoError(t, WriteReport(target.Path, target.Format, sampleCoverage()),
			"writing %s to %s", target.Format, target.Path)
	}

	xml, err := os.ReadFile(stem + ".xml")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(xml), "<?xml"), "cobertura output should be XML")

	info, err := os.ReadFile(stem + ".info")
	require.NoError(t, err)
	assert.Contains(t, string(info), "end_of_record", "lcov output should have records")

	html, err := os.ReadFile(stem + ".html")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(html), "<!doctype html>"), "html output should look like HTML")

	js, err := os.ReadFile(stem + ".json")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(js)), "{"), "json output should look like JSON")
}
