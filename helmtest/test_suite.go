package helmtest

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

var failSuiteLabel = color.New(color.FgWhite, color.BgRed).Sprint(" Fail ")
var passSuiteLabel = color.New(color.FgBlack, color.BgGreen).Sprint(" Pass ")
var emphasize = color.New(color.FgWhite).SprintFunc()

type TestSuite struct {
	Name      string `yaml:"suite"`
	Templates []string
	Tests     []TestJob
	testFile  string
}

func ParseTestSuiteFile(path string) (TestSuite, error) {
	var suite TestSuite
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return TestSuite{}, err
	}

	if err := yaml.Unmarshal(content, &suite); err != nil {
		return TestSuite{}, err
	}

	return suite, nil
}

func (s TestSuite) Run(targetChart *chart.Chart) TestSuiteResult {
	preparedChart, err := s.prepareChart(targetChart)
	if err != nil {
		return TestSuiteResult{ExecError: err}
	}

	suitePass := true
	jobResults := make([]TestJobResult, len(s.Tests))
	for idx, testJob := range s.Tests {
		if len(s.Templates) > 0 {
			testJob.defaultFile = s.Templates[0]
		}
		jobResult := testJob.Run(preparedChart)
		jobResults[idx] = jobResult
		if !jobResult.Passed {
			suitePass = false
		}
	}
	return TestSuiteResult{
		Passed:      suitePass,
		TestsResult: jobResults,
	}
}

func (s TestSuite) printResult(writer io.Writer, pass bool, testsResult string) {
	var label string
	if pass {
		label = passSuiteLabel
	} else {
		label = failSuiteLabel
	}
	var testFilePath string
	if s.testFile != "" {
		testFilePath = path.Dir(s.testFile) +
			string(os.PathSeparator) +
			emphasize(path.Base(s.testFile))
	}
	fmt.Fprintf(writer, "%s %s %s\n", label, emphasize(s.Name), testFilePath)
	fmt.Fprint(writer, testsResult)
}

func (s TestSuite) prepareChart(targetChart *chart.Chart) (*chart.Chart, error) {
	copiedChart := new(chart.Chart)
	copiedChart = targetChart

	if len(s.Templates) > 0 {
		filteredTemplate := make([]*chart.Template, len(s.Templates))
		for idx, fileName := range s.Templates {
			found := false
			for _, template := range targetChart.Templates {
				if path.Base(template.Name) == fileName {
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
