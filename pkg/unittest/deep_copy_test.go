package unittest_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/lrills/helm-unittest/pkg/unittest"
	"github.com/stretchr/testify/assert"
	v3loader "helm.sh/helm/v3/pkg/chart/loader"
)

// Validate the copy of V3Chart with its original.
// func ValidateV3Chart(t *testing.T, targetChart, copiedChart *v3chart.Chart) bool {
// 	result := false

// 	// Check metadata fields

// 	// Recreate the dependencies
// 	// Filter trough dependencies.
// 	for _, dependency := range targetChart.Dependencies() {
// 		copiedChartRoute := filepath.Join(chartRoute, subchartPrefix, dependency.Name())
// 		copiedDependency := CopyV3Chart(copiedChartRoute, templatesToAssert, dependency)
// 		copiedChart.AddDependency(copiedDependency)
// 	}

// 	return copiedChart
// }

func TestCopyHelmChartWithSubSubCharts(t *testing.T) {
	templateAsserts := []string{"**"}

	// Load the chart used by this suite (with logging temporarily disabled)
	log.SetOutput(ioutil.Discard)
	initialChart, _ := v3loader.Load(testV3GlobalDoubleChart)
	log.SetOutput(os.Stdout)

	// Copy
	sut := CopyV3Chart(initialChart.Name(), templateAsserts, initialChart)

	// Validate loaded chart
	assert.NotNil(t, sut)
}
