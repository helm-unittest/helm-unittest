package unittest

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/lrills/helm-unittest/pkg/unittest/validators"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"

	yaml "gopkg.in/yaml.v2"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"
	v2util "k8s.io/helm/pkg/chartutil"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
	v2renderutil "k8s.io/helm/pkg/renderutil"
	v2timeconv "k8s.io/helm/pkg/timeconv"
)

const templatePrefix string = "templates"
const subchartPrefix string = "charts"
const noValueContent string = "<no value>"

// getTemplateFileName,
// Validate if prefix templates is not there,
// used for backward compatibility of old unittests.
func getTemplateFileName(fileName string) string {
	if !strings.HasPrefix(fileName, templatePrefix) && !strings.HasPrefix(fileName, subchartPrefix) {
		// Within templates unix separators are always used.
		return filepath.ToSlash(filepath.Join(templatePrefix, fileName))
	}
	return fileName
}

func spliteChartRoutes(routePath string) []string {
	splited := strings.Split(routePath, string(filepath.Separator))
	routes := make([]string, len(splited)/2+1)
	for r := 0; r < len(routes); r++ {
		routes[r] = splited[r*2]
	}
	return routes
}

func scopeValuesWithRoutes(routes []string, values map[interface{}]interface{}) map[interface{}]interface{} {
	if len(routes) > 1 {
		return scopeValuesWithRoutes(
			routes[:len(routes)-1],
			map[interface{}]interface{}{
				routes[len(routes)-1]: values,
			},
		)
	}
	return values
}

func parseV2RenderError(errorMessage string) (string, string) {
	// Split the error into several groups.
	// those groups are required to parse the correct value.
	const regexPattern string = "^.+\"(.+)\":(?:.+:)* (.+)$"

	filePath, content := parseRenderError(regexPattern, errorMessage)

	return filePath, content
}

func parseV3RenderError(errorMessage string) (string, string) {
	// Split the error into several groups.
	// those groups are required to parse the correct value.
	// ^.+( |\()(.+):\d+:\d+\)?:(.+:)* (.+)$
	const regexPattern string = "^.+(?: |\\()(.+):\\d+:\\d+\\)?:(?:.+:)* (.+)$"

	filePath, content := parseRenderError(regexPattern, errorMessage)

	return filePath, content
}

func parseRenderError(regexPattern, errorMessage string) (string, string) {
	filePath := ""
	content := noValueContent

	r := regexp.MustCompile(regexPattern)
	result := r.FindStringSubmatch(errorMessage)

	if len(result) == 3 {
		filePath = result[1]
		content = fmt.Sprintf("%s: %s", common.RAW, result[2])
	}

	return filePath, content
}

func parseYamlFile(rendered string) ([]common.K8sManifest, error) {
	decoder := yaml.NewDecoder(strings.NewReader(rendered))
	parsedYamls := make([]common.K8sManifest, 0)

	for {
		parsedYaml := common.K8sManifest{}
		if err := decoder.Decode(parsedYaml); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		if len(parsedYaml) > 0 {
			parsedYamls = append(parsedYamls, parsedYaml)
		}
	}

	return parsedYamls, nil
}

func parseTextFile(rendered string) []common.K8sManifest {
	manifests := make([]common.K8sManifest, 0)
	manifest := make(common.K8sManifest)
	manifest[common.RAW] = rendered

	if len(manifest) > 0 {
		manifests = append(manifests, manifest)
	}
	return manifests
}

type orderedSnapshotComparer struct {
	cache   *snapshot.Cache
	test    string
	counter uint
}

func (s *orderedSnapshotComparer) CompareToSnapshot(content interface{}) *snapshot.CompareResult {
	s.counter++
	return s.cache.Compare(s.test, s.counter, content)
}

// TestJob definition of a test, including values and assertions
type TestJob struct {
	Name          string `yaml:"it"`
	Values        []string
	Set           map[string]interface{}
	Template      string
	Templates     []string
	DocumentIndex *int         `yaml:"documentIndex"`
	Assertions    []*Assertion `yaml:"asserts"`
	Release       struct {
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
	// route indicate which chart in the dependency hierarchy
	// like "parant-chart", "parent-charts/charts/child-chart"
	chartRoute string
	// where the test suite file located
	definitionFile string
	// list of templates assertion should assert if not specified
	defaultTemplatesToAssert []string
	// requireSuccess
	requireRenderSuccess bool
}

// RunV2 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV2(
	targetChart *v2chart.Chart,
	cache *snapshot.Cache,
	failfast bool,
	result *results.TestJobResult,
) *results.TestJobResult {
	startTestRun := time.Now()
	t.polishAssertionsTemplate(targetChart.Metadata.Name)
	result.DisplayName = t.Name

	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, renderSucceed, err := t.renderV2Chart(targetChart, userValues)
	if err != nil {
		result.ExecError = err
		return result
	}

	manifestsOfFiles, err := t.parseManifestsFromOutputOfFiles(outputOfFiles)
	if err != nil {
		result.ExecError = err
		return result
	}

	snapshotComparer := &orderedSnapshotComparer{cache: cache, test: t.Name}
	result.Passed, result.AssertsResult = t.runAssertions(
		manifestsOfFiles,
		snapshotComparer,
		renderSucceed,
		nil,
		failfast,
	)

	result.Duration = time.Since(startTestRun)
	return result
}

// RunV3 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV3(
	targetChart *v3chart.Chart,
	cache *snapshot.Cache,
	failfast bool,
	result *results.TestJobResult,
) *results.TestJobResult {
	startTestRun := time.Now()
	t.polishAssertionsTemplate(targetChart.Name())
	result.DisplayName = t.Name

	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, renderSucceed, renderError := t.renderV3Chart(targetChart, userValues)
	if renderError != nil {
		result.ExecError = renderError
		// Continue to enable matching error via failedTemplate assert
	}

	manifestsOfFiles, err := t.parseManifestsFromOutputOfFiles(outputOfFiles)
	if err != nil {
		result.ExecError = err
		return result
	}

	snapshotComparer := &orderedSnapshotComparer{cache: cache, test: t.Name}
	result.Passed, result.AssertsResult = t.runAssertions(
		manifestsOfFiles,
		snapshotComparer,
		renderSucceed,
		renderError,
		failfast,
	)

	result.Duration = time.Since(startTestRun)
	return result
}

// liberally borrows from helm-template
func (t *TestJob) getUserValues() ([]byte, error) {
	base := map[interface{}]interface{}{}
	routes := spliteChartRoutes(t.chartRoute)

	for _, specifiedPath := range t.Values {
		value := map[interface{}]interface{}{}
		var valueFilePath string
		if path.IsAbs(specifiedPath) {
			valueFilePath = specifiedPath
		} else {
			valueFilePath = filepath.Join(filepath.Dir(t.definitionFile), specifiedPath)
		}

		bytes, err := ioutil.ReadFile(valueFilePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &value); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", specifiedPath, err)
		}
		base = valueutils.MergeValues(base, scopeValuesWithRoutes(routes, value))
	}

	for path, values := range t.Set {
		setMap, err := valueutils.BuildValueOfSetPath(values, path)
		if err != nil {
			return []byte{}, err
		}

		base = valueutils.MergeValues(base, scopeValuesWithRoutes(routes, setMap))
	}
	return yaml.Marshal(base)
}

// render the V2chart and return result map
func (t *TestJob) renderV2Chart(targetChart *v2chart.Chart, userValues []byte) (map[string]string, bool, error) {
	renderSucceed := true
	config := &v2chart.Config{Raw: string(userValues), Values: map[string]*v2chart.Value{}}
	kubeVersion := fmt.Sprintf("%s.%s", v2util.DefaultKubeVersion.Major, v2util.DefaultKubeVersion.Minor)

	if t.Capabilities.MajorVersion != "" && t.Capabilities.MinorVersion != "" {
		kubeVersion = fmt.Sprintf("%s.%s", t.Capabilities.MajorVersion, t.Capabilities.MinorVersion)
	}

	renderOpts := v2renderutil.Options{
		ReleaseOptions: *t.releaseV2Option(),
		KubeVersion:    kubeVersion,
		APIVersions:    t.Capabilities.APIVersions,
	}

	// Override the chart version when version is setup in test.
	if t.Chart.Version != "" {
		targetChart.Metadata.Version = t.Chart.Version
	}

	// Override the chart appVErsion when version is setup in test.
	if t.Chart.AppVersion != "" {
		targetChart.Metadata.AppVersion = t.Chart.AppVersion
	}

	// Filter the files that needs to be validated
	filteredChart := t.filterV2Chart(targetChart)

	outputOfFiles, err := v2renderutil.Render(filteredChart, config, renderOpts)
	// When rendering failed, due to fail or required,
	// make sure to translate the error to outputOfFiles.
	if err != nil {
		renderSucceed = false
		// Parse the error and create an outputFile
		filePath, content := parseV2RenderError(err.Error())
		// If error not parsed well, rethrow as normal.
		if filePath == "" {
			return nil, renderSucceed, err
		}
		outputOfFiles[filePath] = content
	}

	return outputOfFiles, renderSucceed, nil
}

// Filter the V2Chart and its dependencies with partials and selected test files.
func (t *TestJob) filterV2Chart(targetChart *v2chart.Chart) *v2chart.Chart {
	copiedChart := new(v2chart.Chart)
	*copiedChart = *targetChart

	suiteIsFromRootChart := len(strings.Split(t.chartRoute, string(filepath.Separator))) <= 1
	dependencyChart := ""

	if len(t.defaultTemplatesToAssert) == 0 && suiteIsFromRootChart {
		return copiedChart
	}

	if suiteIsFromRootChart {
		copiedChart.Templates = t.filterV2Templates(t.chartRoute, dependencyChart, targetChart)
	}

	// Filter trough dependencies.
	filteredDependencies := make([]*v2chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies {
		if suiteIsFromRootChart {
			dependencyChart = dependency.Metadata.Name
		}

		filteredDependencyTemplates := t.filterV2Templates(t.chartRoute, dependencyChart, dependency)
		if len(filteredDependencyTemplates) > 0 {
			copiedDependencyChart := new(v2chart.Chart)
			*copiedDependencyChart = *dependency
			copiedDependencyChart.Templates = filteredDependencyTemplates
			filteredDependencies = append(filteredDependencies, copiedDependencyChart)
		}
	}
	copiedChart.Dependencies = filteredDependencies

	return copiedChart
}

// filterV2Templates, Filter the V2Templates with only the partials and selected test files.
func (t *TestJob) filterV2Templates(chartRoute, dependecyChart string, targetChart *v2chart.Chart) []*v2chart.Template {
	filteredV2Template := make([]*v2chart.Template, 0)

	for _, fileName := range t.defaultTemplatesToAssert {
		for _, template := range targetChart.Templates {
			selectedV2TemplateName := filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName)))
			foundV2TemplateName := filepath.ToSlash(filepath.Join(chartRoute, template.Name))

			if dependecyChart != "" {
				foundV2TemplateName = filepath.ToSlash(filepath.Join(chartRoute, "charts", dependecyChart, template.Name))
			}

			if foundV2TemplateName == selectedV2TemplateName {
				filteredV2Template = append(filteredV2Template, template)
				break
			}
		}
	}

	// add partial templates
	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			filteredV2Template = append(filteredV2Template, template)
		}
	}

	return filteredV2Template
}

// render the chart and return result map
func (t *TestJob) renderV3Chart(targetChart *v3chart.Chart, userValues []byte) (map[string]string, bool, error) {
	renderSucceed := true
	values, err := v3util.ReadValues(userValues)
	if err != nil {
		return nil, false, err
	}
	options := *t.releaseV3Option()

	//Check Release Name length
	if t.Release.Name != "" {
		err = v3util.ValidateReleaseName(t.Release.Name)
		if err != nil {
			return nil, false, err
		}
	}

	// Override the chart version when version is setup in test.
	if t.Chart.Version != "" {
		targetChart.Metadata.Version = t.Chart.Version
	}

	// Override the chart appVErsion when version is setup in test.
	if t.Chart.AppVersion != "" {
		targetChart.Metadata.AppVersion = t.Chart.AppVersion
	}

	err = v3util.ProcessDependencies(targetChart, values)
	if err != nil {
		return nil, false, err
	}

	vals, err := v3util.ToRenderValues(targetChart, values.AsMap(), options, t.capabilitiesV3())
	if err != nil {
		return nil, false, err
	}

	// Filter the files that needs to be validated
	filteredChart := t.filterV3Chart(targetChart)

	outputOfFiles, err := v3engine.Render(filteredChart, vals)

	// When rendering failed, due to fail or required,
	// make sure to translate the error to outputOfFiles.
	if err != nil {
		// When no failed assertion is set, the error can be send directly as a failure.
		if t.requireRenderSuccess {
			return nil, false, err
		}

		renderSucceed = false
		// Parse the error and create an outputFile
		filePath, content := parseV3RenderError(err.Error())
		// If error not parsed well, rethrow as normal.
		if filePath == "" && content != noValueContent {
			return nil, renderSucceed, err
		}

		// If error, validate if template error occurred
		if strings.HasPrefix(filepath.Base(filePath), "_") {
			for _, fileName := range t.defaultTemplatesToAssert {
				selectedTemplateName := filepath.ToSlash(filepath.Join(t.chartRoute, getTemplateFileName(fileName)))
				outputOfFiles[selectedTemplateName] = content
			}
		} else {
			outputOfFiles[filePath] = content
		}
	}

	return outputOfFiles, renderSucceed, nil
}

// Filter the V3Chart and its dependencies with partials and selected test files.
func (t *TestJob) filterV3Chart(targetChart *v3chart.Chart) *v3chart.Chart {
	copiedChart := new(v3chart.Chart)
	*copiedChart = *targetChart

	suiteIsFromRootChart := len(strings.Split(t.chartRoute, string(filepath.Separator))) <= 1
	dependencyChart := ""

	if len(t.defaultTemplatesToAssert) == 0 && suiteIsFromRootChart {
		return copiedChart
	}

	if suiteIsFromRootChart {
		copiedChart.Templates = t.filterV3Templates(t.chartRoute, dependencyChart, targetChart)
	}

	// Filter trough dependencies.
	filteredDependencies := make([]*v3chart.Chart, 0)
	for _, dependency := range targetChart.Dependencies() {
		if suiteIsFromRootChart {
			dependencyChart = dependency.Metadata.Name
		}

		filteredDependencyTemplates := t.filterV3Templates(t.chartRoute, dependencyChart, dependency)
		if len(filteredDependencyTemplates) > 0 {
			copiedDependencyChart := new(v3chart.Chart)
			*copiedDependencyChart = *dependency
			copiedDependencyChart.Templates = filteredDependencyTemplates
			filteredDependencies = append(filteredDependencies, copiedDependencyChart)
		}
	}
	copiedChart.SetDependencies(filteredDependencies...)

	return copiedChart
}

// filterV3Templates, Filter the V3Templates with only the partials and selected test files.
func (t *TestJob) filterV3Templates(chartRoute, dependecyChart string, targetChart *v3chart.Chart) []*v3chart.File {
	filteredV3Template := make([]*v3chart.File, 0)
	// check templates in chart
	for _, fileName := range t.defaultTemplatesToAssert {
		for _, template := range targetChart.Templates {
			selectedV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, getTemplateFileName(fileName)))
			foundV3TemplateName := filepath.ToSlash(filepath.Join(chartRoute, template.Name))

			if dependecyChart != "" {
				foundV3TemplateName = filepath.ToSlash(filepath.Join(chartRoute, "charts", dependecyChart, template.Name))
			}

			if foundV3TemplateName == selectedV3TemplateName {
				filteredV3Template = append(filteredV3Template, template)
				break
			}
		}
	}

	// add partial templates
	for _, template := range targetChart.Templates {
		if strings.HasPrefix(filepath.Base(template.Name), "_") {
			filteredV3Template = append(filteredV3Template, template)
		}
	}

	return filteredV3Template
}

// get chartutil.ReleaseOptions ready for render
func (t *TestJob) releaseV2Option() *v2util.ReleaseOptions {
	options := v2util.ReleaseOptions{
		Name:      "RELEASE-NAME",
		Namespace: "NAMESPACE",
		Time:      v2timeconv.Now(),
		Revision:  t.Release.Revision,
		IsInstall: !t.Release.IsUpgrade,
		IsUpgrade: t.Release.IsUpgrade,
	}
	if t.Release.Name != "" {
		options.Name = t.Release.Name
	}
	if t.Release.Namespace != "" {
		options.Namespace = t.Release.Namespace
	}
	return &options
}

// get chartutil.ReleaseOptions ready for render
func (t *TestJob) releaseV3Option() *v3util.ReleaseOptions {
	options := v3util.ReleaseOptions{
		Name:      "RELEASE-NAME",
		Namespace: "NAMESPACE",
		Revision:  t.Release.Revision,
		IsInstall: !t.Release.IsUpgrade,
		IsUpgrade: t.Release.IsUpgrade,
	}
	if t.Release.Name != "" {
		options.Name = t.Release.Name
	}
	if t.Release.Namespace != "" {
		options.Namespace = t.Release.Namespace
	}
	return &options
}

// get chartutil.Capabilities ready for render
func (t *TestJob) capabilitiesV3() *v3util.Capabilities {
	capabilities := v3util.DefaultCapabilities

	// Override the version, when set.
	if t.Capabilities.MajorVersion != "" && t.Capabilities.MinorVersion != "" {
		capabilities.KubeVersion = v3util.KubeVersion{
			Version: fmt.Sprintf("v%s.%s.0", t.Capabilities.MajorVersion, t.Capabilities.MinorVersion),
			Major:   t.Capabilities.MajorVersion,
			Minor:   t.Capabilities.MinorVersion,
		}
	}

	// Add ApiVersions when set
	capabilities.APIVersions = v3util.VersionSet(t.Capabilities.APIVersions)

	return capabilities
}

// parse rendered manifest if it's yaml
func (t *TestJob) parseManifestsFromOutputOfFiles(outputOfFiles map[string]string) (
	map[string][]common.K8sManifest,
	error,
) {
	manifestsOfFiles := make(map[string][]common.K8sManifest)

	for file, rendered := range outputOfFiles {

		switch filepath.Ext(file) {
		case ".yaml":
			manifest, err := parseYamlFile(rendered)
			if err != nil {
				return nil, err
			}
			manifestsOfFiles[file] = manifest
		case ".txt":
			manifestsOfFiles[file] = parseTextFile(rendered)
		}

	}

	return manifestsOfFiles, nil
}

// run Assert of all assertions of test
func (t *TestJob) runAssertions(
	manifestsOfFiles map[string][]common.K8sManifest,
	snapshotComparer validators.SnapshotComparer,
	renderSucceed bool, renderError error, failfast bool,
) (bool, []*results.AssertionResult) {
	testPass := true
	assertsResult := make([]*results.AssertionResult, 0)

	for idx, assertion := range t.Assertions {
		result := assertion.Assert(
			manifestsOfFiles,
			snapshotComparer,
			renderSucceed,
			renderError,
			&results.AssertionResult{Index: idx},
		)

		assertsResult = append(assertsResult, result)
		testPass = testPass && result.Passed

		if !testPass && failfast {
			break
		}
	}
	return testPass, assertsResult
}

// add prefix to Assertion.Template
func (t *TestJob) polishAssertionsTemplate(targetChartName string) {
	if t.chartRoute == "" {
		t.chartRoute = targetChartName
	}

	t.requireRenderSuccess = true

	for _, assertion := range t.Assertions {
		t.requireRenderSuccess = t.requireRenderSuccess && assertion.requireRenderSuccess
		templatesToAssert := make([]string, 0)

		if t.DocumentIndex != nil {
			assertion.DocumentIndex = *t.DocumentIndex
		}

		if assertion.Template == "" {
			if len(t.Templates) > 0 {
				templatesToAssert = append(templatesToAssert, t.Templates...)
			} else if t.Template == "" {
				templatesToAssert = t.defaultTemplatesToAssert
			} else {
				templatesToAssert = append(templatesToAssert, t.Template)
			}
		} else {
			templatesToAssert = append(templatesToAssert, assertion.Template)
		}

		// map the file name to the path of helm rendered result
		templatesPath := make([]string, 0)
		for _, template := range templatesToAssert {
			templatePath := filepath.ToSlash(filepath.Join(t.chartRoute, getTemplateFileName(template)))
			templatesPath = append(templatesPath, templatePath)
		}
		assertion.defaultTemplates = templatesPath
	}
}
