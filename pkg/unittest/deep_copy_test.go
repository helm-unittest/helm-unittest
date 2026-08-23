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
	v2chart "helm.sh/helm/v4/pkg/chart/v2"
	v2loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

func templatesCount(targetChart *v2chart.Chart) int {
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
	initialChart, _ := v2loader.Load(testV4GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartWithSubChartsNoFilter(t *testing.T) {
	templateAsserts := []string{"**"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 58, templatesCount)
}

func TestFullCopyHelmChartWithSubCharts(t *testing.T) {
	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := FullCopyV2Chart(initialChart.Name(), initialChart.Name(), initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 58, templatesCount)
}

func TestCopyHelmChartSingleChartSpecialFilenames(t *testing.T) {
	templateAsserts := []string{"*.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithFilesChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestFullCopyHelmChartSingleChartSpecialFilenames(t *testing.T) {
	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithFilesChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := FullCopyV2Chart(initialChart.Name(), initialChart.Name(), initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubChartInRootDeployment(t *testing.T) {
	templateAsserts := []string{"charts/postgresql/templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubSubChartInRootDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithSubSubFolderChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartSingleSubChartInSubChartDeployment(t *testing.T) {
	templateAsserts := []string{"templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WithSubChart)
	log.SetOutput(os.Stdout)
	chartRoute := filepath.Join(initialChart.Name(), "charts", "child-chart")

	// Copy
	sut := CopyV2Chart(chartRoute, initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 1, templatesCount)
}

func TestCopyHelmChartWithSubSubChartsAllConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"**/configmap.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 3, templatesCount)
}

func TestCopyHelmChartWithSubSubChartsRootchartConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"*/configmap.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 2, templatesCount)
}

func TestCopyHelmChartWithSamenameSubSubChartsConfigMapFilter(t *testing.T) {
	templateAsserts := []string{"charts/with-samenamesubsubcharts/charts/with-samenamesubsubcharts/templates/deployment.yaml"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(io.Discard)
	initialChart, _ := v2loader.Load(testV4WitSamenameSubSubChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, []string{}, initialChart)

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
	initialChart, _ := v2loader.Load(testV4BasicChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV2Chart(initialChart.Name(), initialChart.Name(), templateAsserts, excludedTemplateAsserts, initialChart)

	templatesCount := templatesCount(sut)

	// Validate loaded chart
	assert.NotNil(t, sut)
	assert.Equal(t, 10, templatesCount)
}
