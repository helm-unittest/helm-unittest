package unittest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"gopkg.in/yaml.v3"
	v3loader "helm.sh/helm/v3/pkg/chart/loader"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"

	log "github.com/sirupsen/logrus"
)

// m modifier: multi line. Causes ^ and $ to match the begin/end of each line (not only begin/end of string)
var splitterPattern = regexp.MustCompile("(?m:^---$)")

// ParseTestSuiteFile parse a suite file that contain one or more suites at path and returns an array of TestSuite
func ParseTestSuiteFile(suiteFilePath, chartRoute string, strict bool, valueFilesSet []string) ([]*TestSuite, error) {
	content, err := os.ReadFile(suiteFilePath)
	if err != nil {
		return []*TestSuite{{chartRoute: chartRoute}}, err
	}

	// The pattern matches lines that contain only three hyphens (---), which is a common
	// delimiter used in various file formats (e.g., YAML, Markdown) to separate sections.
	// The -1 passed as the third argument to Split tells it to return all parts,
	// including the parts matched by the regular expression pattern.
	parts := splitterPattern.Split(string(content), -1)
	log.WithField(common.LOG_TEST_SUITE, "parse-test-suite-file").Debug("suite '", suiteFilePath, "' total parts ", len(parts))
	var testSuites []*TestSuite
	for _, part := range parts {
		// Ensure the part has data exclude whitespace, otherwise we can ignore the split
		if len(strings.TrimSpace(part)) > 0 {
			testSuite, suiteErr := createTestSuite(suiteFilePath, chartRoute, part, strict, valueFilesSet, false)
			testSuites = append(testSuites, testSuite)
			if suiteErr != nil {
				return testSuites, suiteErr
			}
		}
	}

	return testSuites, nil
}

func createTestSuite(suiteFilePath string, chartRoute string, content string, strict bool, valueFilesSet []string, fromRender bool) (*TestSuite, error) {
	suite := TestSuite{
		chartRoute: chartRoute,
		fromRender: fromRender,
	}

	var err error
	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(suiteFilePath)
	suite.definitionFile, err = filepath.Rel(cwd, absPath)
	if err != nil {
		return &suite, err
	}

	// Use decoder to setup strict or unstrict
	yamlDecoder := yaml.NewDecoder(strings.NewReader(content))
	yamlDecoder.KnownFields(strict)
	if err := yamlDecoder.Decode(&suite); err != nil {
		return &suite, err
	}

	err = suite.validateTestSuite()
	if err != nil {
		return &suite, err
	}

	// Append the value files from command to the test suites.
	suite.Values = append(suite.Values, valueFilesSet...)

	return &suite, nil
}

// RenderTestSuiteFiles renders a helm suite of test files and returns their TestSuites
func RenderTestSuiteFiles(helmTestSuiteDir string, chartRoute string, strict bool, valueFilesSet []string, renderValues map[string]interface{}) ([]*TestSuite, error) {
	testChartPath := filepath.Join(helmTestSuiteDir, "Chart.yaml")

	// Ensure there's a helm file
	if _, err := os.Stat(testChartPath); err != nil {
		return nil, err
	}

	chart, err := v3loader.Load(helmTestSuiteDir)
	if err != nil {
		return nil, err
	}

	options := v3util.ReleaseOptions{
		Name:      "TEST-SUITE-RELEASE",
		Namespace: "NAMESPACE",
		Revision:  1,
		IsInstall: false,
		IsUpgrade: false,
	}

	values, err := v3util.ToRenderValues(chart, renderValues, options, nil)
	if err != nil {
		return nil, err
	}
	renderedFiles, err := v3engine.Render(chart, values)
	if err != nil {
		return nil, err
	}

	// Iterate over all keys
	// Split any multiple suites
	// Filter any empty templates
	// Set up a numerical snapshot idx if none provided
	// Check that we didn't make a bunch of empty yamls
	renderErrs, suites := iterateAllKeys(renderedFiles, chart.Name(), helmTestSuiteDir, chartRoute, strict, valueFilesSet)

	if len(renderErrs) > 0 {
		return nil, errors.Join(renderErrs...)
	}

	return suites, nil
}

func iterateAllKeys(renderedFiles map[string]string, chartName, helmTestSuiteDir, chartRoute string, strict bool, valueFilesSet []string) ([]error, []*TestSuite) {
	renderErrs := make([]error, 0)
	suites := make([]*TestSuite, 0)

	for templateName, template := range renderedFiles {
		if len(strings.TrimSpace(template)) == 0 {
			renderErrs = append(renderErrs, fmt.Errorf("test suite template (%s) file did not render a manifest", templateName))
			continue
		}

		templateFilePath := strings.Replace(templateName, chartName, "", 1)
		absPath := filepath.Join(helmTestSuiteDir, templateFilePath)

		var subYamlErrs []error
		var previousSuitesLen int
		subYamlErrs, previousSuitesLen, suites = iterateTemplates(template, suites, absPath, chartRoute, strict, valueFilesSet)
		if len(subYamlErrs) > 0 {
			renderErrs = append(renderErrs, fmt.Errorf("test suite template (%s) error: %w", templateName, errors.Join(subYamlErrs...)))
		}

		if previousSuitesLen == len(suites) {
			renderErrs = append(renderErrs, fmt.Errorf("test suite template (%s) file did not render a manifest", templateName))
		}
	}
	return renderErrs, suites
}

func iterateTemplates(template string, suites []*TestSuite, absPath string, chartRoute string, strict bool, valueFilesSet []string) ([]error, int, []*TestSuite) {
	var subYamlErrs []error
	templates := splitterPattern.Split(template, -1)
	previousSuitesLen := len(suites)
	realIdx := -1
	for idx, subYaml := range templates {
		if len(strings.TrimSpace(subYaml)) == 0 {
			continue
		}
		realIdx++

		suite, err := createTestSuite(absPath, chartRoute, subYaml, strict, valueFilesSet, true)
		if err != nil {
			subYamlErrs = append(subYamlErrs, fmt.Errorf("chart %d error: %w", idx, err))
			continue
		}

		if len(suite.SnapshotId) == 0 {
			suite.SnapshotId = fmt.Sprintf("%d", realIdx)
		}
		suites = append(suites, suite)
	}
	return subYamlErrs, previousSuitesLen, suites
}

// TestSuite defines scope and templates to render and tests to run
type TestSuite struct {
	Name      string `yaml:"suite"`
	Values    []string
	Set       map[string]interface{}
	Templates []string
	Release   struct {
		Name      string
		Namespace string
		Revision  int
		IsUpgrade bool `yaml:"upgrade"`
	}
	Chart struct {
		Version    string
		AppVersion string `yaml:"appVersion"`
	}
	Capabilities struct {
		MajorVersion string   `yaml:"majorVersion"`
		MinorVersion string   `yaml:"minorVersion"`
		APIVersions  []string `yaml:"apiVersions"`
	}
	KubernetesProvider KubernetesFakeClientProvider `yaml:"kubernetesProvider"`
	Tests              []*TestJob
	// where the test suite file located
	definitionFile string
	// route indicate which chart in the dependency hierarchy
	// like "parent-chart", "parent-charts/charts/child-chart"
	chartRoute string
	// if true, indicates that this was created from a helm rendered file
	fromRender bool
	// An identifier to append to snapshot files
	SnapshotId string `yaml:"snapshotId"`
}

// RunV3 runs all the test jobs defined in TestSuite.
func (s *TestSuite) RunV3(
	chartPath string,
	snapshotCache *snapshot.Cache,
	failfast bool,
	renderPath string,
	result *results.TestSuiteResult,
) *results.TestSuiteResult {
	s.polishTestJobsPathInfo()

	result.DisplayName = s.Name
	result.FilePath = s.definitionFile

	result.Passed, result.TestsResult = s.runV3TestJobs(
		chartPath,
		snapshotCache,
		failfast,
		renderPath,
	)

	result.CountSnapshot(snapshotCache)
	return result
}

// fill file path related info of TestJob
func (s *TestSuite) polishTestJobsPathInfo() {
	log.WithField(common.LOG_TEST_SUITE, "polish-test-jobs-path-info").Debug("suite '", s.Name, "' total tests ", len(s.Tests))
	for _, test := range s.Tests {
		test.chartRoute = s.chartRoute
		test.definitionFile = s.definitionFile

		s.polishReleaseSettings(test)
		s.polishCapabilitiesSettings(test)
		s.polishKubernetesProviderSettings(test)
		s.polishChartSettings(test)

		// Make deep clone of global set
		test.globalSet = copySet(s.Set)
		if len(s.Values) > 0 {
			test.Values = append(s.Values, test.Values...)
		}
		log.WithField(common.LOG_TEST_SUITE, "polish-test-jobs-path-info").Debug("test '", test.Name, "' with total values ", len(test.Values), " and ", test.Values)

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

func (s *TestSuite) polishKubernetesProviderSettings(test *TestJob) {
	if len(s.KubernetesProvider.Objects) > 0 {
		test.KubernetesProvider.Objects = append(test.KubernetesProvider.Objects, s.KubernetesProvider.Objects...)
	}
	if len(s.KubernetesProvider.Scheme) > 0 {
		if test.KubernetesProvider.Scheme == nil {
			test.KubernetesProvider.Scheme = map[string]KubernetesFakeKindProps{}
		}
		for k, v := range s.KubernetesProvider.Scheme {
			test.KubernetesProvider.Scheme[k] = v
		}
	}
}

// override chart settings in testjobs when defined in testsuite
func (s *TestSuite) polishChartSettings(test *TestJob) {
	if test.Chart.Version == "" && s.Chart.Version != "" {
		test.Chart.Version = s.Chart.Version
	}
	if test.Chart.AppVersion == "" && s.Chart.AppVersion != "" {
		test.Chart.AppVersion = s.Chart.AppVersion
	}
}

func (s *TestSuite) runV3TestJobs(
	chartPath string,
	cache *snapshot.Cache,
	failfast bool,
	renderPath string,
) (bool, []*results.TestJobResult) {
	suitePass := false
	jobResults := make([]*results.TestJobResult, len(s.Tests))

	for idx, testJob := range s.Tests {
		// (Re)load the chart used by this suite (with logging temporarily disabled)
		log.SetOutput(io.Discard)
		chart, _ := v3loader.Load(chartPath)
		log.SetOutput(os.Stdout)

		jobResult := testJob.RunV3(chart, cache, failfast, renderPath, &results.TestJobResult{Index: idx})
		jobResults[idx] = jobResult

		if idx == 0 {
			suitePass = jobResult.Passed
		}

		suitePass = suitePass && jobResult.Passed

		if !suitePass && failfast {
			break
		}
	}
	return suitePass, jobResults
}

func (s *TestSuite) validateTestSuite() error {
	if len(s.Tests) == 0 {
		return fmt.Errorf("no tests found")
	}

	if s.fromRender && len(s.Name) == 0 {
		return fmt.Errorf(("helm chart based test suites must include `suite` field"))
	}

	for _, testJob := range s.Tests {
		if len(testJob.Assertions) == 0 {
			log.WithField(common.LOG_TEST_SUITE, "validate-test-suite").Debugln("no asserts found", testJob)
			return fmt.Errorf("no asserts found")
		}
	}

	return nil
}

func (s *TestSuite) SnapshotFileUrl() string {
	if len(s.SnapshotId) > 0 {
		// appedn the snapshot id
		return fmt.Sprintf("%s_%s", s.definitionFile, s.SnapshotId)
	}
	return s.definitionFile
}
