package unittest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/stretchr/testify/assert"
)

func filesHelper(t *testing.T) {
	t.Helper()
	testPath, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(testPath)
	})
}

func assertArrayPathOsAgnostic(t *testing.T, expected, actual []string) {
	t.Helper()
	var want []string
	for _, el := range expected {
		// required as Linux separator is '/' when Windows is '\\'
		want = append(want, filepath.FromSlash(el))
	}
	assert.Equal(t, want, actual)
}

func TestGetFiles_ChartWithoutSubCharts(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/basic")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assert.Equal(t, len(actual), 17)
}

func TestGetFiles_ChartWithoutSubChartsNoDuplicates(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/basic")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/configmap_test.yaml", "tests/configmap_test.yaml", "tests/configmap_test.yaml"}, false)
	assert.NoError(t, err)

	assert.Equal(t, len(actual), 1)
	assertArrayPathOsAgnostic(t, []string{"tests/configmap_test.yaml"}, actual)
}

func TestGetFiles_ChartWithoutSubChartsTopLevel(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3")
	assert.NoError(t, err)

	actual, err := GetFiles("basic", []string{"tests/configmap_test.yaml", "tests/not-exists.yaml"}, false)
	assert.NoError(t, err)

	assert.Equal(t, len(actual), 1)
	assertArrayPathOsAgnostic(t, []string{"basic/tests/configmap_test.yaml"}, actual)
}

func TestGetFiles_ChartWithSubChartCdToSubChart(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles("charts/child-chart", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(actual))
	assertArrayPathOsAgnostic(t, []string{
		"charts/child-chart/tests/child_chart_test.yaml",
		"charts/child-chart/tests/deployment_test.yaml",
		"charts/child-chart/tests/hpa_test.yaml",
		"charts/child-chart/tests/ingress_test.yaml",
		"charts/child-chart/tests/notes_test.yaml",
		"charts/child-chart/tests/service_test.yaml",
	}, actual)
}

func TestGetFiles_ChartWithSubChartFromRootDefaultPattern(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assertArrayPathOsAgnostic(t, []string{
		"tests/all-charts_test.yaml",
		"tests/certmanager_test.yaml",
		"tests/deployment_test.yaml",
		"tests/ingress_test.yaml",
		"tests/notes_test.yaml",
		"tests/postgresql_deployment_test.yaml",
		"tests/postgresql_secrets_test.yaml",
		"tests/service_test.yaml",
	}, actual)
}

func TestGetFiles_ChartWithSubChartFromRootVisibleSubChartTests(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"charts/child-chart/tests/deployment_test.yaml"}, false)
	assert.NoError(t, err)
	assertArrayPathOsAgnostic(t, []string{"charts/child-chart/tests/deployment_test.yaml"}, actual)
}

func TestGetFiles_ChartWithSubChartPatternMatchingParentAndSubChart(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	pattern := []string{"tests/deployment_test.yaml"}

	parent, err := GetFiles(".", []string{"tests/deployment_test.yaml"}, false)
	assert.NoError(t, err)
	subchart, err := GetFiles("charts/child-chart", pattern, false)
	assert.NoError(t, err)

	actual := append(parent, subchart...)

	assertArrayPathOsAgnostic(t, []string{
		"tests/deployment_test.yaml",
		"charts/child-chart/tests/deployment_test.yaml",
	}, actual)
}

func TestGetFiles_ChartWithSubChartPatternMatchingChildTests(t *testing.T) {
	filesHelper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	pattern := []string{"charts/child-chart/tests/deployment_test.yaml"}

	parent, err := GetFiles(".", pattern, false)
	assert.NoError(t, err)
	subchart, err := GetFiles("charts/child-chart", pattern, false)
	assert.NoError(t, err)

	actual := append(parent, subchart...)

	// Pattern found when executing from parent and child charts
	expected := []string{
		"charts/child-chart/tests/deployment_test.yaml",
		"charts/child-chart/tests/deployment_test.yaml",
	}

	if runtime.GOOS == "windows" {
		expected = []string{
			"charts\\child-chart\\tests\\deployment_test.yaml",
			"charts/child-chart/tests/deployment_test.yaml",
		}
	}

	assert.Equal(t, expected, actual)
}

func TestWithDifferentPatterns(t *testing.T) {
	tmp := t.TempDir()
	path := fmt.Sprintf("%s/%s", tmp, "./a/b/c/e")
	err := os.MkdirAll(path, 0755)
	assert.NoError(t, err)

	fs := fstest.MapFS{
		"a/b/c/first.yaml":     {Data: []byte("hi")},
		"a/b/c/e/second.yaml":  {Data: []byte("hi")},
		"a/b/third.yaml":       {Data: []byte("hi")},
		"a/b/c/third.hcl":      {Data: []byte("hi")},
		"a/b/c/file0.txt":      {Data: []byte("hi")},
		"a/file1.txt":          {Data: []byte("hi")},
		"a/b/c/file2.json":     {Data: []byte("hi")},
		"a/b/file2.json":       {Data: []byte("hi")},
		"file3.xml":            {Data: []byte("hi")},
		"a/b/file4_test.csv":   {Data: []byte("hi")},
		"a/b/c/e/file5.ini":    {Data: []byte("hi")},
		"a/b/c/e/file6.log":    {Data: []byte("hi")},
		"a/b/c/e/file7.conf":   {Data: []byte("hi")},
		"a/b/file6.md":         {Data: []byte("hi")},
		"a/b/c/e/file8.md":     {Data: []byte("hi")},
		"a/b/c/e/file9.html":   {Data: []byte("hi")},
		"a/b/c/e/file10.css":   {Data: []byte("hi")},
		"a/b/c/file11.js":      {Data: []byte("hi")},
		"a/b/file12_test.go":   {Data: []byte("hi")},
		"a/b/c/e/file13.py":    {Data: []byte("hi")},
		"a/file14_test.rb":     {Data: []byte("hi")},
		"a/b/c/file15.php":     {Data: []byte("hi")},
		"a/b/file16.sh":        {Data: []byte("hi")},
		"a/b/c/e/file17.pl":    {Data: []byte("hi")},
		"a/b/c/e/file18.rs":    {Data: []byte("hi")},
		"a/b/c/e/file19.kt":    {Data: []byte("hi")},
		"a/b/c/e/file20.swift": {Data: []byte("hi")},
	}

	for path, el := range fs {
		err := os.WriteFile(filepath.Join(tmp, path), el.Data, 0644)
		assert.NoError(t, err)
	}

	tests := []struct {
		pattern     []string
		expected    []string
		skipWindows bool
	}{
		{
			pattern: []string{"**/*.yaml"},
			expected: []string{
				fmt.Sprintf("%s/a/b/c/first.yaml", tmp),
				fmt.Sprintf("%s/a/b/c/e/second.yaml", tmp),
				fmt.Sprintf("%s/a/b/third.yaml", tmp),
			},
		},
		{
			pattern:  []string{"[a-z\\.\\/]*"},
			expected: []string{},
		},
		{
			pattern: []string{"/[a-z\\.\\/]*"},
			expected: []string{
				"/[a-z\\.\\/]*",
			},
			skipWindows: true,
		},
		{
			pattern: []string{"**/*.log"},
			expected: []string{
				fmt.Sprintf("%s/a/b/c/e/file6.log", tmp),
			},
		},
		{
			pattern:  []string{"**/*.(json|js|log)"},
			expected: []string{},
		},
		{
			pattern: []string{"a/*/file*.json", "a/**/file*.txt"},
			expected: []string{
				fmt.Sprintf("%s/a/b/file2.json", tmp),
				fmt.Sprintf("%s/a/file1.txt", tmp),
				fmt.Sprintf("%s/a/b/c/file0.txt", tmp),
			},
		},
		{
			pattern: []string{"**/*.xml", "**/*.csv"},
			expected: []string{
				fmt.Sprintf("%s/file3.xml", tmp),
				fmt.Sprintf("%s/a/b/file4_test.csv", tmp),
			},
		},
		{
			pattern:  []string{"a/b/c/e/**.md"},
			expected: []string{},
		},
		{
			pattern: []string{"a/b/c/e/*.md"},
			expected: []string{
				fmt.Sprintf("%s/a/b/c/e/file8.md", tmp),
			},
		},
		{
			pattern: []string{"a/*_test.rb", "a/b/**.rb", "a/b/c/**.php"},
			expected: []string{
				fmt.Sprintf("%s/a/file14_test.rb", tmp),
			},
		},
		{
			pattern: []string{"a/b/*_test.go", ".*11.js"},
			expected: []string{
				fmt.Sprintf("%s/a/b/file12_test.go", tmp),
			},
		},
		{
			pattern: []string{fmt.Sprintf("%s/a/b/*.sh", tmp), "**/*.pl"},
			expected: []string{
				fmt.Sprintf("%s/a/b/*.sh", tmp),
				fmt.Sprintf("%s/a/b/c/e/file17.pl", tmp),
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("with %s identified %d files", strings.Join(tt.pattern, ":"), len(tt.expected)), func(t *testing.T) {
			if tt.skipWindows && runtime.GOOS == "windows" {
				t.Skip("Skip this test on Windows")
			}
			var escapePattern []string
			for _, p := range tt.pattern {
				escapePattern = append(escapePattern, filepath.FromSlash(p))
			}
			files, _ := GetFiles(tmp, escapePattern, false)
			assert.Equal(t, len(tt.expected), len(files))
			for _, expected := range tt.expected {
				assert.Contains(t, files, filepath.FromSlash(expected))
			}
		})
	}
}

func TestGetFiles_GlobError(t *testing.T) {
	tmp := t.TempDir()

	path := fmt.Sprintf("%s/%s", tmp, "./a/b/c.d/e.f")
	err := os.MkdirAll(path, 0755)
	assert.NoError(t, err)

	files, err := GetFiles(path, []string{"[**"}, false)
	assert.Nil(t, files)
	assert.Error(t, err)
	assert.EqualError(t, err, "syntax error in pattern")
}
