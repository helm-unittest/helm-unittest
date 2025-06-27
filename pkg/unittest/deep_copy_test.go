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

func TestCopyHelmChartSingleDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartWithSubChartsNoFilter(t *testing.T) {
	templateAsserts := []string{"**"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 58, templatesCount)
}

func TestCopyHelmChartSingleChartSpecialFilenames(t *testing.T) {
	templateAsserts := []string{"*.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithFilesChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

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
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubSubChartInRootDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WithSubSubFolderChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

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
	chartRoute := filepath.Join(initialChart.Name(), "charts", "child-chart")

	// Copy
	sut := CopyV3Chart(chartRoute, initialChart.Name(), templateAsserts, []string{}, initialChart)

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
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

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
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 2, templatesCount)
}

func TestCopyHelmChartWithSamenameSubSubChartsConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"charts/with-samenamesubsubcharts/charts/with-samenamesubsubcharts/templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3WitSamenameSubSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartWithExcludedTemplatesFilter(t *testing.T) {
	templateAsserts := []string{"*.yaml"}
	excludedTemplateAsserts := []string{"deployment.yaml", "ing*"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v3loader.Load(testV3BasicChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), initialChart.Name(), templateAsserts, excludedTemplateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 10, templatesCount)
}
