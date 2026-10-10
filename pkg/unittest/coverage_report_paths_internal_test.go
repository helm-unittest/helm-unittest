package unittest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: report file paths (Cobertura filename, LCOV SF:) must resolve
// relative to the chart path passed on the CLI, not the chart's declared
// Chart.yaml name, so Codecov/SonarQube/GitLab can find the source file.
func TestV4RunnerCoverageReportPathsResolveFromChartPath(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		Coverage:     true,
		WithSubChart: true,
		TestFiles:    []string{"tests/*_test.yaml"},
	}

	chartPath := "../../test/data/v3/coverage-fromjson"
	passed := runner.RunV4([]string{chartPath})
	require.True(t, passed, buffer.String())
	require.Len(t, runner.coverageReports, 1)

	require.NotEmpty(t, runner.coverageReports[0].Files)
	for _, f := range runner.coverageReports[0].Files {
		assert.True(t, strings.HasPrefix(f.Name, chartPath+"/"),
			"report path %q should be rooted at the chart path %q passed on the CLI", f.Name, chartPath)
	}
}
