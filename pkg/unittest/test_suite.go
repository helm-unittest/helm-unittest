package unittest

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"gopkg.in/yaml.v2"
	v3chart "helm.sh/helm/v3/pkg/chart"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
)

// ParseTestSuiteFile parse a suite file at path and returns TestSuite
func ParseTestSuiteFile(suiteFilePath, chartRoute string, strict bool, valueFilesSet []string) (*TestSuite, error) {
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

	if strict {
		if err := yaml.UnmarshalStrict(content, &suite); err != nil {
			return &suite, err
		}
	} else {
		if err := yaml.Unmarshal(content, &suite); err != nil {
			return &suite, err
		}
	}

	// Append the valuesfiles from command to the testsuites.
	suite.Values = append(suite.Values, valueFilesSet...)

	return &suite, nil
}

// TestSuite defines scope and templates to render and tests to run
type TestSuite struct {
	Name      string `yaml:"suite"`
	Values    []string
	Templates []string
	Release   struct {
		Name      string
		Namespace string
		Revision  int
		IsUpgrade bool `yaml:"upgrade"`
	}
	Chart struct {
		Version    string
		AppVersion string
	}
	Capabilities struct {
		MajorVersion string   `yaml:"majorVersion"`
		MinorVersion string   `yaml:"minorVersion"`
		APIVersions  []string `yaml:"apiVersions"`
	}
	Tests []*TestJob
	// where the test suite file located
	definitionFile string
	// route indicate which chart in the dependency hierarchy
	// like "parent-chart", "parent-charts/charts/child-chart"
	chartRoute string
}

// RunV2 runs all the test jobs defined in TestSuite.
func (s *TestSuite) RunV2(
	targetChart *v2chart.Chart,
	snapshotCache *snapshot.Cache,
	failfast bool,
	result *results.TestSuiteResult,
) *results.TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	result.Passed, result.TestsResult = s.runV2TestJobs(
		targetChart,
		snapshotCache,
		failfast,
	)

	result.CountSnapshot(snapshotCache)
	return result
}

// RunV3 runs all the test jobs defined in TestSuite.
func (s *TestSuite) RunV3(
	targetChart *v3chart.Chart,
	snapshotCache *snapshot.Cache,
	failfast bool,
	result *results.TestSuiteResult,
) *results.TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	result.Passed, result.TestsResult = s.runV3TestJobs(
		targetChart,
		snapshotCache,
		failfast,
	)

	result.CountSnapshot(snapshotCache)
	return result
}

// fill file path related info of TestJob
func (s *TestSuite) polishTestJobsPathInfo() {
	for _, test := range s.Tests {
		test.chartRoute = s.chartRoute
		test.definitionFile = s.definitionFile

		s.polishReleaseSettings(test)
		s.polishCapabilitiesSettings(test)
		s.polishChartSettings(test)

		if len(s.Values) > 0 {
			test.Values = append(test.Values, s.Values...)
		}

		if len(s.Templates) > 0 {
			test.defaultTemplatesToAssert = s.Templates
		}
	}
}

// override release settings in testjobs when defined in testsuite
func (s *TestSuite) polishReleaseSettings(test *TestJob) {
	if s.Release.Name != "" {
		if test.Release.Name == "" {
			test.Release.Name = s.Release.Name
		}
	}

	if s.Release.Namespace != "" {
		if test.Release.Namespace == "" {
			test.Release.Namespace = s.Release.Namespace
		}
	}

	if s.Release.Revision > 0 {
		if test.Release.Revision == 0 {
			test.Release.Revision = s.Release.Revision
		}
	}

	if s.Release.IsUpgrade {
		if !test.Release.IsUpgrade {
			test.Release.IsUpgrade = s.Release.IsUpgrade
		}
	}
}

// override capabilities settings in testjobs when defined in testsuite
func (s *TestSuite) polishCapabilitiesSettings(test *TestJob) {
	if s.Capabilities.MajorVersion != "" && s.Capabilities.MinorVersion != "" {
		if test.Capabilities.MajorVersion == "" && test.Capabilities.MinorVersion == "" {
			test.Capabilities.MajorVersion = s.Capabilities.MajorVersion
			test.Capabilities.MinorVersion = s.Capabilities.MinorVersion
		}
	}

	if len(s.Capabilities.APIVersions) > 0 {
		test.Capabilities.APIVersions = append(test.Capabilities.APIVersions, s.Capabilities.APIVersions...)
	}
}

// override chart settings in testjobs when defined in testsuite
func (s *TestSuite) polishChartSettings(test *TestJob) {
	if s.Chart.Version != "" {
		test.Chart.Version = s.Chart.Version
	}
	if s.Chart.AppVersion != "" {
		test.Chart.AppVersion = s.Chart.AppVersion
	}
}

func (s *TestSuite) runV2TestJobs(
	chart *v2chart.Chart,
	cache *snapshot.Cache,
	failfast bool,
) (bool, []*results.TestJobResult) {
	suitePass := true
	jobResults := make([]*results.TestJobResult, len(s.Tests))
	dependenciesBackup := make([]*v2chart.Chart, len(chart.Dependencies))
	copy(dependenciesBackup, chart.Dependencies)

	for idx, testJob := range s.Tests {
		jobResult := testJob.RunV2(chart, cache, failfast, &results.TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if !jobResult.Passed {
			suitePass = false
		}

		chart.Dependencies = dependenciesBackup

		if !suitePass && failfast {
			break
		}
	}
	return suitePass, jobResults
}

func (s *TestSuite) runV3TestJobs(
	chart *v3chart.Chart,
	cache *snapshot.Cache,
	failfast bool,
) (bool, []*results.TestJobResult) {
	suitePass := true
	jobResults := make([]*results.TestJobResult, len(s.Tests))
	metadataDependenciesBackup := cloneDependencies(chart.Metadata.Dependencies)
	dependenciesBackup := chart.Dependencies()
	valuesBackup := cloneValues(chart.Values)

	for idx, testJob := range s.Tests {
		jobResult := testJob.RunV3(chart, cache, failfast, &results.TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if !jobResult.Passed {
			suitePass = false
		}

		chart.SetDependencies(dependenciesBackup...)
		chart.Values = nil
		chart.Values = cloneValues(valuesBackup)
		chart.Metadata.Dependencies = nil
		chart.Metadata.Dependencies = cloneDependencies(metadataDependenciesBackup)

		if !suitePass && failfast {
			break
		}
	}
	return suitePass, jobResults
}

func cloneDependencies(metadataDependencies []*v3chart.Dependency) []*v3chart.Dependency {
	clonedDependencies := make([]*v3chart.Dependency, 0)

	for _, metadataDependency := range metadataDependencies {
		clonedDependency := &v3chart.Dependency{
			Name:         metadataDependency.Name,
			Version:      metadataDependency.Version,
			Repository:   metadataDependency.Repository,
			Condition:    metadataDependency.Condition,
			Tags:         metadataDependency.Tags,
			Enabled:      metadataDependency.Enabled,
			ImportValues: cloneImportValues(metadataDependency.ImportValues),
			Alias:        metadataDependency.Alias,
		}

		clonedDependencies = append(clonedDependencies, clonedDependency)
	}

	return clonedDependencies
}

func cloneValues(values map[string]interface{}) map[string]interface{} {
	clonedValues := make(map[string]interface{})

	for key, value := range values {
		clonedValues[key] = value
	}

	return clonedValues
}

func cloneImportValues(importValues []interface{}) []interface{} {
	clonedImportValues := make([]interface{}, 0)

	clonedImportValues = append(clonedImportValues, importValues...)

	return clonedImportValues
}
