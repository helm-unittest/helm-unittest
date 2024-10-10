package unittest_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/stretchr/testify/assert"
)

func helper(t *testing.T) {
	t.Helper()
	testPath, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(testPath)
	})
}

func assertResults(t *testing.T, expected, actual []string) {
	t.Helper()
	var want []string
	for _, el := range expected {
		// required as Linux separator is '/' when Windows is '\\'
		want = append(want, filepath.FromSlash(el))
	}
	assert.Equal(t, want, actual)
}

func TestGetFiles_ChartWithoutSubCharts(t *testing.T) {
	helper(t)
	err := os.Chdir("../../test/data/v3/basic")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assert.Equal(t, len(actual), 11)
}

func TestGetFiles_ChartWithoutSubChartsNoDuplicates(t *testing.T) {
	helper(t)
	err := os.Chdir("../../test/data/v3/basic")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/configmap_test.yaml", "tests/configmap_test.yaml", "tests/configmap_test.yaml"}, false)
	assert.NoError(t, err)

	assert.Equal(t, len(actual), 1)
	assertResults(t, []string{"tests/configmap_test.yaml"}, actual)
}

func TestGetFiles_ChartWithoutSubChartsTopLevel(t *testing.T) {
	helper(t)
	err := os.Chdir("../../test/data/v3")
	assert.NoError(t, err)

	actual, err := GetFiles("basic", []string{"tests/configmap_test.yaml", "tests/not-exists.yaml"}, false)
	assert.NoError(t, err)

	assert.Equal(t, len(actual), 1)
	assertResults(t, []string{"basic/tests/configmap_test.yaml"}, actual)
}

func TestGetFiles_ChartWithSubChartCdToSubChart(t *testing.T) {
	helper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles("charts/child-chart", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(actual))
	assertResults(t, []string{
		"charts/child-chart/tests/child_chart_test.yaml",
		"charts/child-chart/tests/deployment_test.yaml",
		"charts/child-chart/tests/hpa_test.yaml",
		"charts/child-chart/tests/ingress_test.yaml",
		"charts/child-chart/tests/notes_test.yaml",
		"charts/child-chart/tests/service_test.yaml",
	}, actual)
}

func TestGetFiles_ChartWithSubChartFromRootDefaultPattern(t *testing.T) {
	helper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"tests/*_test.yaml"}, false)
	assert.NoError(t, err)
	assertResults(t, []string{
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
	helper(t)
	err := os.Chdir("../../test/data/v3/with-subchart")
	assert.NoError(t, err)

	actual, err := GetFiles(".", []string{"charts/child-chart/tests/deployment_test.yaml"}, false)
	assert.NoError(t, err)
	assertResults(t, []string{"charts/child-chart/tests/deployment_test.yaml"}, actual)
}
