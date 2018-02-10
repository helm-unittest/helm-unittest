package helmtest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type TestSuite struct {
	Name      string `yaml:"suite"`
	Templates []string
	Tests     []TestJob
	filePath  string
}

func ParseTestSuiteFile(path string) (TestSuite, error) {
	var suite TestSuite
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return TestSuite{}, err
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(path)
	suite.filePath, err = filepath.Rel(cwd, absPath)
	if err != nil {
		return TestSuite{}, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return TestSuite{}, err
	}

	return suite, nil
}

func (s TestSuite) Run(targetChart *chart.Chart, result *TestSuiteResult) *TestSuiteResult {
	result.DisplayName = s.Name

	preparedChart, err := s.prepareChart(targetChart)
	if err != nil {
		result.ExecError = err
		return result
	}

	suitePass := true
	jobResults := make([]*TestJobResult, len(s.Tests))
	for idx, testJob := range s.Tests {
		if len(s.Templates) > 0 {
			testJob.defaultFile = s.Templates[0]
		}
		jobResult := testJob.Run(preparedChart, &TestJobResult{Index: idx})
		jobResults[idx] = jobResult
		if !jobResult.Passed {
			suitePass = false
		}
	}
	result.Passed = suitePass
	result.TestsResult = jobResults
	return result
}

func (s TestSuite) prepareChart(targetChart *chart.Chart) (*chart.Chart, error) {
	copiedChart := new(chart.Chart)
	copiedChart = targetChart

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
