package unittest

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"

	yaml "gopkg.in/yaml.v3"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"
)

func spliteChartRoutes(routePath string) []string {
	splited := strings.Split(routePath, string(filepath.Separator))
	routes := make([]string, len(splited)/2+1)
	for r := 0; r < len(routes); r++ {
		routes[r] = splited[r*2]
	}
	return routes
}

func scopeValuesWithRoutes(routes []string, values map[string]interface{}) map[string]interface{} {
	if len(routes) > 1 {
		return scopeValuesWithRoutes(
			routes[:len(routes)-1],
			map[string]interface{}{
				routes[len(routes)-1]: values,
			},
		)
	}
	return values
}

func parseV3RenderError(errorMessage string) (string, map[string]string) {
	// Split the error into several groups.
	// those groups are required to parse the correct value.
	// ^.+( |\()(.+):\d+:\d+\)?:(.+:)* (.+)$
	// (?mU)^.+(?: |\\()(.+):\\d+:\\d+\\)?:(?:.+:)* (.+)$
	// (?mU)^(?:.+: |.+ \()(?:(.+):\d+:\d+).+(?:.+>)*: (.+)$
	const regexPattern string = "(?mU)^(?:.+: |.+ \\()(?:(.+):\\d+:\\d+).+(?:.+>)*: (.+)$"

	filePath, content := parseRenderError(regexPattern, errorMessage)

	return filePath, content
}

func parseRenderError(regexPattern, errorMessage string) (string, map[string]string) {
	filePath := ""
	content := map[string]string{
		common.RAW: "",
	}

	r := regexp.MustCompile(regexPattern)
	result := r.FindStringSubmatch(errorMessage)

	if len(result) == 3 {
		filePath = result[1]
		content[common.RAW] = result[2]
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
	DocumentIndex *int `yaml:"documentIndex"`
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
	Assertions []*Assertion `yaml:"asserts"`

	// global set values
	globalSet map[string]interface{}
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

// RunV3 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV3(
	targetChart *v3chart.Chart,
	cache *snapshot.Cache,
	failfast bool,
	result *results.TestJobResult,
) *results.TestJobResult {
	startTestRun := time.Now()
	t.determineRenderSuccess()
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

	// Setup Assertion Templates based on the chartname and outputOfFiles
	t.polishAssertionsTemplate(targetChart.Name(), outputOfFiles)

	manifestsOfFiles, err := t.parseManifestsFromOutputOfFiles(targetChart.Name(), outputOfFiles)
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
	base := map[string]interface{}{}
	routes := spliteChartRoutes(t.chartRoute)

	for _, specifiedPath := range t.Values {
		value := map[string]interface{}{}
		var valueFilePath string
		if path.IsAbs(specifiedPath) {
			valueFilePath = specifiedPath
		} else {
			valueFilePath = filepath.Join(filepath.Dir(t.definitionFile), specifiedPath)
		}

		bytes, err := os.ReadFile(valueFilePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &value); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", specifiedPath, err)
		}
		base = valueutils.MergeValues(base, scopeValuesWithRoutes(routes, value))
	}

	// Merge global set values before merging the other set values
	for path, values := range t.globalSet {
		setMap, err := valueutils.BuildValueOfSetPath(values, path)
		if err != nil {
			return []byte{}, err
		}

		base = valueutils.MergeValues(base, scopeValuesWithRoutes(routes, setMap))
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

// render the chart and return result map
func (t *TestJob) renderV3Chart(targetChart *v3chart.Chart, userValues []byte) (map[string]string, bool, error) {
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

	// When defaultTemplatesToAssert is empty, ensure all templates will be validated.
	if len(t.defaultTemplatesToAssert) == 0 {
		// Set all files
		t.defaultTemplatesToAssert = []string{multiWildcard}
	}

	// Filter the files that needs to be validated
	filteredChart := CopyV3Chart(targetChart.Name(), t.defaultTemplatesToAssert, targetChart)

	outputOfFiles, err := v3engine.Render(filteredChart, vals)

	var renderSucceed bool
	outputOfFiles, renderSucceed, err = t.translateErrorToOutputFiles(err, outputOfFiles)
	if err != nil {
		return nil, false, err
	}

	return outputOfFiles, renderSucceed, nil
}

// When rendering failed, due to fail or required,
// make sure to translate the error to outputOfFiles.
func (t *TestJob) translateErrorToOutputFiles(err error, outputOfFiles map[string]string) (map[string]string, bool, error) {
	renderSucceed := true
	if err != nil {
		renderSucceed = false
		// When no failed assertion is set, the error can be send directly as a failure.
		if t.requireRenderSuccess {
			return nil, renderSucceed, err
		}

		// Parse the error and create an outputFile
		filePath, content := parseV3RenderError(err.Error())
		// If error not parsed well, rethrow as normal.
		if filePath == "" && len(content[common.RAW]) == 0 {
			return nil, renderSucceed, err
		}

		// If error, validate if template error occurred
		if strings.HasPrefix(filepath.Base(filePath), "_") {
			for _, fileName := range t.defaultTemplatesToAssert {
				selectedTemplateName := filepath.ToSlash(filepath.Join(t.chartRoute, getTemplateFileName(fileName)))
				outputOfFiles[selectedTemplateName] = common.TrustedMarshalYAML(content)
			}
		} else {
			outputOfFiles[filePath] = common.TrustedMarshalYAML(content)
		}
	}

	return outputOfFiles, renderSucceed, nil
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
func (t *TestJob) parseManifestsFromOutputOfFiles(targetChartName string, outputOfFiles map[string]string) (
	map[string][]common.K8sManifest,
	error,
) {
	manifestsOfFiles := make(map[string][]common.K8sManifest)

	for file, rendered := range outputOfFiles {
		if !strings.HasPrefix(file, targetChartName) {
			file = filepath.ToSlash(filepath.Join(targetChartName, file))
		}

		switch filepath.Ext(file) {
		case ".yaml", ".yml", ".tpl":
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
	testPass := false
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

		if idx == 0 {
			testPass = result.Passed
		}

		testPass = testPass && result.Passed

		if !testPass && failfast {
			break
		}
	}
	return testPass, assertsResult
}

// determine if the success for rendering is required,
// to return an errorCode direct.
func (t *TestJob) determineRenderSuccess() {
	t.requireRenderSuccess = true

	for _, assertion := range t.Assertions {
		t.requireRenderSuccess = t.requireRenderSuccess && assertion.requireRenderSuccess
	}
}

// add prefix to Assertion.Template
func (t *TestJob) polishAssertionsTemplate(targetChartName string, outputOfFiles map[string]string) {
	if t.chartRoute == "" {
		t.chartRoute = targetChartName
	}

	for _, assertion := range t.Assertions {
		prefixedChartsNameFiles := false
		templatesToAssert := make([]string, 0)

		if t.DocumentIndex != nil {
			assertion.DocumentIndex = *t.DocumentIndex
		}

		if assertion.Template == "" {
			if len(t.Templates) > 0 {
				templatesToAssert = append(templatesToAssert, t.Templates...)
			} else if t.Template == "" {
				templatesToAssert, prefixedChartsNameFiles = t.resolveDefaultTemplatesToAssert(outputOfFiles)
			} else {
				templatesToAssert = append(templatesToAssert, t.Template)
			}
		} else {
			templatesToAssert = append(templatesToAssert, assertion.Template)
		}

		// map the file name to the path of helm rendered result
		assertion.defaultTemplates = t.prefixTemplatesToAssert(templatesToAssert, prefixedChartsNameFiles)
	}
}

func (t *TestJob) resolveDefaultTemplatesToAssert(outputOfFiles map[string]string) ([]string, bool) {
	defaultTemplatesPath := make([]string, 0)
	resetAsserts := false

	for _, template := range t.defaultTemplatesToAssert {
		if strings.Contains(template, "*") {
			resetAsserts = true
			break
		}
	}

	if resetAsserts {
		for template := range outputOfFiles {
			defaultTemplatesPath = append(defaultTemplatesPath, template)
		}
	} else {
		defaultTemplatesPath = t.defaultTemplatesToAssert
	}

	return defaultTemplatesPath, resetAsserts
}

func (t *TestJob) prefixTemplatesToAssert(templatesToAssert []string, prefixedChartsNameFiles bool) []string {
	templatesPath := make([]string, 0)

	if !prefixedChartsNameFiles {
		for _, template := range templatesToAssert {
			templatePath := filepath.ToSlash(filepath.Join(t.chartRoute, getTemplateFileName(template)))
			templatesPath = append(templatesPath, templatePath)
		}
	} else {
		templatesPath = templatesToAssert
	}

	return templatesPath
}
