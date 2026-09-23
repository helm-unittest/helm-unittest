package unittest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/coverage"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: an earlier suite that renders coverage with a subchart disabled must not corrupt the shared instrumented chart used by a later suite's coverage render.
func TestV4RunnerCoverageIsolatedAcrossSuites(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		Coverage:     true,
		WithSubChart: true,
		TestFiles:    []string{"tests/*_test.yaml"},
	}

	passed := runner.RunV4([]string{"../../test/data/v3/coverage-subchart"})
	require.True(t, passed, buffer.String())
	require.Len(t, runner.coverageReports, 1)

	byName := func(suffix string) (coverage.FileCoverage, bool) {
		for _, f := range runner.coverageReports[0].Files {
			if strings.HasSuffix(f.Name, suffix) {
				return f, true
			}
		}
		return coverage.FileCoverage{}, false
	}

	childCM, ok := byName("charts/child/templates/cm.yaml")
	require.True(t, ok, "child subchart template missing from coverage report")
	assert.True(t, childCM.Rendered, "child subchart template lost its Rendered flag across suites")
	assert.Positive(t, childCM.Branches.Covered, "child subchart branch coverage lost across suites")

	parentUsesChild, ok := byName("templates/parent-uses-child.yaml")
	require.True(t, ok, "parent template missing from coverage report")
	assert.True(t, parentUsesChild.Rendered, "parent template that includes a child define lost coverage across suites")
	assert.Positive(t, parentUsesChild.Actions.Covered, "parent template action coverage lost across suites")
}

// Regression: the coverage render must resolve `lookup` against the test's KubernetesProvider, or lookup-dependent branches diverge from the primary render and coverage is lost.
func TestV4RunnerCoverageResolvesLookups(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		Coverage:     true,
		WithSubChart: true,
		TestFiles:    []string{"tests/*_test.yaml"},
	}

	passed := runner.RunV4([]string{"../../test/data/v3/with-k8s-fake-client"})
	require.True(t, passed, buffer.String())
	require.Len(t, runner.coverageReports, 1)

	totals := runner.coverageReports[0].Totals.Actions
	assert.Positive(t, totals.Total)
	assert.Equal(t, totals.Total, totals.Covered, "lookup-dependent templates must be fully credited in the coverage render")
}

// Regression: a define whose output is consumed as data via `include ... | fromJson` must have its own branches credited without corrupting the parsed JSON.
func TestV4RunnerCoverageCreditsFromJsonHelper(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:      printer.NewPrinter(buffer, nil),
		Coverage:     true,
		WithSubChart: true,
		TestFiles:    []string{"tests/*_test.yaml"},
	}

	passed := runner.RunV4([]string{"../../test/data/v3/coverage-fromjson"})
	require.True(t, passed, buffer.String())
	require.Len(t, runner.coverageReports, 1)

	var helper coverage.FileCoverage
	found := false
	for _, f := range runner.coverageReports[0].Files {
		if strings.HasSuffix(f.Name, "templates/_helpers.tpl") {
			helper = f
			found = true
		}
	}
	require.True(t, found, "fromJson-consumed helper missing from coverage report")
	assert.True(t, helper.Rendered, "fromJson-consumed helper must be credited")
	assert.Equal(t, 2, helper.Branches.Total)
	assert.Equal(t, 2, helper.Branches.Covered, "both branches of the fromJson-consumed helper must be credited")
}
