package unittest

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
	Name           string `yaml:"suite"`
	Templates      []string
	Tests          []*TestJob
	definitionFile string
}

func ParseTestSuiteFile(path string) (*TestSuite, error) {
	var suite TestSuite
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return &suite, err
	}

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(path)
	suite.definitionFile, err = filepath.Rel(cwd, absPath)
	if err != nil {
		return &suite, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return &suite, err
	}

	return &suite, nil
}

func (s *TestSuite) Run(targetChart *chart.Chart, result *TestSuiteResult) *TestSuiteResult {
	result.DisplayName = s.Name
	result.FilePath = s.definitionFile
	s.polishTestJob()

	preparedChart, err := s.prepareChart(targetChart)
	if err != nil {
		result.ExecError = err
		return result
	}

	result.Passed, result.TestsResult = s.runTestJobs(preparedChart)
	return result
}

func (s *TestSuite) polishTestJob() {
	for _, test := range s.Tests {
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

func (s *TestSuite) runTestJobs(chart *chart.Chart) (bool, []*TestJobResult) {
	suitePass := true
	jobResults := make([]*TestJobResult, len(s.Tests))
	for idx, testJob := range s.Tests {
		jobResult := testJob.Run(chart, &TestJobResult{Index: idx})
		jobResults[idx] = jobResult
		if !jobResult.Passed {
			suitePass = false
		}
	}
	return suitePass, jobResults
}
