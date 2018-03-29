package unittest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/lrills/helm-unittest/unittest/snapshot"
	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// ParseTestSuiteFile parse a suite file at path and returns TestSuite
func ParseTestSuiteFile(suiteFilePath, chartRoute string) (*TestSuite, error) {
	suite := TestSuite{chartRoute: chartRoute}
	content, err := ioutil.ReadFile(suiteFilePath)
	if err != nil {
		return &suite, err
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(suiteFilePath)
	suite.definitionFile, err = filepath.Rel(cwd, absPath)
	if err != nil {
		return &suite, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return &suite, err
	}

	return &suite, nil
}

// TestSuite defines scope and templates to render and tests to run
type TestSuite struct {
	Name      string `yaml:"suite"`
	Templates []string
	Tests     []*TestJob
	// where the test suite file located
	definitionFile string
	// route indicate which chart in the dependency hierarchy
	// like "parant-chart", "parent-charts/charts/child-chart"
	chartRoute string
}

// Run runs all the test jobs defined in TestSuite
func (s *TestSuite) Run(
	targetChart *chart.Chart,
	snapshotCache *snapshot.Cache,
	result *TestSuiteResult,
) *TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	preparedChart, err := s.prepareChart(targetChart)
	if err != nil {
		result.ExecError = err
		return result
	}

	result.Passed, result.TestsResult = s.runTestJobs(
		preparedChart,
		snapshotCache,
	)

	countSnapshot(result, snapshotCache)
	return result
}

// fill file path related info of TestJob
func (s *TestSuite) polishTestJobsPathInfo() {
	for _, test := range s.Tests {
		test.chartRoute = s.chartRoute
		test.definitionFile = s.definitionFile
		if len(s.Templates) > 0 {
			test.defaultTemplateToAssert = s.Templates[0]
		}
	}
}

func (s *TestSuite) prepareChart(targetChart *chart.Chart) (*chart.Chart, error) {
	copiedChart := new(chart.Chart)
	*copiedChart = *targetChart

	if len(s.Templates) > 0 {
		filteredTemplate := make([]*chart.Template, len(s.Templates))
		for idx, fileName := range s.Templates {
			found := false
			for _, template := range targetChart.Templates {
				if filepath.Base(template.Name) == fileName {
					filteredTemplate[idx] = template
					found = true
					break
				}
			}
			if !found {
				return &chart.Chart{}, fmt.Errorf(
					"template file `templates/%s` not found in chart",
					fileName,
				)
			}
		}

		for _, template := range targetChart.Templates {
			if path.Ext(template.Name) == ".tpl" {
				filteredTemplate = append(filteredTemplate, template)
			}
		}
		copiedChart.Templates = filteredTemplate
	}
	return copiedChart, nil
}

func (s *TestSuite) runTestJobs(
	chart *chart.Chart,
	cache *snapshot.Cache,
) (bool, []*TestJobResult) {
	suitePass := true
	jobResults := make([]*TestJobResult, len(s.Tests))

	for idx, testJob := range s.Tests {
		jobResult := testJob.Run(chart, cache, &TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if !jobResult.Passed {
			suitePass = false
		}
	}
	return suitePass, jobResults
}

func countSnapshot(result *TestSuiteResult, cache *snapshot.Cache) {
	result.SnapshotCounting.Created = cache.InsertedCount()
	result.SnapshotCounting.Failed = cache.FailedCount()
	result.SnapshotCounting.Total = cache.CurrentCount()
	result.SnapshotCounting.Vanished = cache.VanishedCount()
}
