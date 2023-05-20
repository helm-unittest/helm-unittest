package unittest_test

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	v3loader "helm.sh/helm/v3/pkg/chart/loader"
)

func templatesCount(targetChart *chart.Chart) int {
	totalCount := len(targetChart.Templates)

	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			totalCount--
		}
	}

	for _, dependency := range targetChart.Dependencies() {
		totalCount += templatesCount(dependency)
	}

	return totalCount
}

func TestCopyHelmChartWithSubChartsNoFilter(t *testing.T) {
	templateAsserts := []string{}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 14, templatesCount)
}

func TestCopyHelmChartSingleDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubChartInRootDeployment(t *testing.T) {
	templateAsserts := []string{"charts/postgresql/templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubChartInSubChartDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithSubChart)
	log.SetOutput(os.Stdout)
	chartRoute := filepath.ToSlash(filepath.Join(initialChart.Name(), "charts", "postgresql"))

	// Copy
	sut := CopyV3Chart(chartRoute, templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartWithSubSubChartsAllConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"**/configmap.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 3, templatesCount)
}

func TestCopyHelmChartWithSubSubChartsRootchartConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"*/configmap.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 2, templatesCount)
}
