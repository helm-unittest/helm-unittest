package unittest

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"
)

const LOG_TEST_JOB = "test-job"

// Split the error into several groups.
// those groups are required to parse the correct value.
// ^.+( |\()(.+):\d+:\d+\)?:(.+:)* (.+)$
// (?mU)^.+(?: |\\()(.+):\\d+:\\d+\\)?:(?:.+:)* (.+)$
// (?mU)^(?:.+: |.+ \()(?:(.+):\d+:\d+).+(?:.+>)*: (.+)$
// (?msU)
//
//	--- m: Multi-line mode. ^ and $ match the start and end of each line.
//	--- s: Dot-all mode. . matches any character, including newline.
//	--- U: Ungreedy mode. Makes quantifiers lazy by default.
//
const regexPattern string = "(?msU)^(?:.+: |.+ \\()(?:(.+):\\d+:\\d+).+(?:.+>)*: (.+)$"

var regexErrorPattern = regexp.MustCompile(regexPattern)

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
		newvalues := make(map[string]interface{})
		if v, ok := values["global"]; ok {
			newvalues["global"] = v
		}
		newvalues[routes[len(routes)-1]] = values
		return scopeValuesWithRoutes(
			routes[:len(routes)-1],
			newvalues,
		)
	}
	return values
}

func parseV3RenderError(errorMessage string) (string, map[string]string) {
	filePath, content := parseRenderError(errorMessage)
	return filePath, content
}

func parseRenderError(errorMessage string) (string, map[string]string) {
	filePath := ""
	content := map[string]string{
		common.RAW: "",
	}

	result := regexErrorPattern.FindStringSubmatch(errorMessage)

	if len(result) == 3 {
		filePath = result[1]
		// check where or not errorMessage is a multiline error message
		lines := strings.SplitN(errorMessage, "\n", 2)
		if len(lines) > 1 {
			content[common.RAW] = lines[1]
		} else {
			// return error unparsed message
			content[common.RAW] = result[2]
		}
	}

	return filePath, content
}

func parseYamlFile(rendered string) ([]common.K8sManifest, error) {
	// Replace --- with ---\n to ensure yaml rendering is parsed correctly/
	rendered = splitterPattern.ReplaceAllString(rendered, "\n---\n")
	decoder := common.YamlNewDecoder(strings.NewReader(rendered))
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

func writeRenderedOutput(renderPath string, outputOfFiles map[string]string) error {
	if renderPath != "" {
		for file, rendered := range outputOfFiles {
			filePath := filepath.Join(renderPath, file)
			directory := filepath.Dir(filePath)
			if _, dirErr := os.Stat(directory); errors.Is(dirErr, os.ErrNotExist) {
				if createDirErr := os.MkdirAll(directory, 0755); createDirErr != nil {
					return createDirErr
				}
			}
			if createFileErr := os.WriteFile(filePath, []byte(rendered), 0644); createFileErr != nil {
				return createFileErr
			}
		}
	}
	return nil
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

type Capabilities struct {
	MajorVersion string   `yaml:"majorVersion"`
	MinorVersion string   `yaml:"minorVersion"`
	APIVersions  []string `yaml:"apiVersions"`
}

// CapabilitiesFields required to identify where or not the filed is provided, and the value is unset or not
type CapabilitiesFields map[string]interface{}

// TestJob definition of a test, including values and assertions
type TestJob struct {
	Name             string `yaml:"it"`
	Values           []string
	Set              map[string]interface{}
	Template         string
	Templates        []string
	DocumentIndex    *int `yaml:"documentIndex"`
	DocumentIndices  map[string][]int
	DocumentSelector *valueutils.DocumentSelector `yaml:"documentSelector"`
	Release          struct {
		Name      string
		Namespace string
		Revision  int
		IsUpgrade bool `yaml:"upgrade"`
	}
	Chart struct {
		Version    string
		AppVersion string `yaml:"appVersion"`
	}
	Capabilities       Capabilities                 `yaml:"-"`
	CapabilitiesFields CapabilitiesFields           `yaml:"capabilities"`
	Assertions         []*Assertion                 `yaml:"asserts"`
	KubernetesProvider KubernetesFakeClientProvider `yaml:"kubernetesProvider"`
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
	renderPath string,
	result *results.TestJobResult,
) *results.TestJobResult {
	startTestRun := time.Now()
	log.WithField(LOG_TEST_JOB, "run-v3").Debug("job name ", t.Name)
	t.determineRenderSuccess()
	result.DisplayName = t.Name
	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, renderSucceed, renderError := t.renderV3Chart(targetChart, []byte(userValues))
	writeError := writeRenderedOutput(renderPath, outputOfFiles)
	if writeError != nil {
		result.ExecError = writeError
		return result
	}

	if renderError != nil {
		result.ExecError = renderError
		// Continue to enable matching error via failedTemplate assert
	}

	manifestsOfFiles, err := t.parseManifestsFromOutputOfFiles(targetChart.Name(), outputOfFiles)
	if err != nil {
		result.ExecError = err
		return result
	}
	// Setup Assertion Templates based on the chartname, documentIndex and outputOfFiles
	t.polishAssertionsTemplate(targetChart.Name(), outputOfFiles)
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
func (t *TestJob) getUserValues() (string, error) {
	base := map[string]interface{}{}
	routes := spliteChartRoutes(t.chartRoute)

	// Load and merge values files.
	for _, specifiedPath := range t.Values {
		value := map[string]interface{}{}
		var valueFilePath string
		if filepath.IsAbs(specifiedPath) {
			valueFilePath = specifiedPath
		} else {
			valueFilePath = filepath.Join(filepath.Dir(t.definitionFile), specifiedPath)
		}

		bytes, err := os.ReadFile(valueFilePath)
		if err != nil {
			return "", err
		}

		if err := common.YmlUnmarshal(string(bytes), &value); err != nil {
			return "", fmt.Errorf("failed to parse %s: %s", specifiedPath, err)
		}

		base = v3util.MergeTables(scopeValuesWithRoutes(routes, value), base)
	}

	// Merge global set values before merging the other set values
	for path, values := range t.globalSet {
		setMap, err := valueutils.BuildValueOfSetPath(values, path)
		if err != nil {
			return "", err
		}

		base = v3util.MergeTables(scopeValuesWithRoutes(routes, setMap), base)
	}

	for path, values := range t.Set {
		setMap, err := valueutils.BuildValueOfSetPath(values, path)
		if err != nil {
			return "", err
		}

		base = v3util.MergeTables(scopeValuesWithRoutes(routes, setMap), base)
	}
	log.WithField(LOG_TEST_JOB, "get-user-values").Debug("values ", base)
	return common.YmlMarshall(base)
}

// render the chart and return result map
func (t *TestJob) renderV3Chart(targetChart *v3chart.Chart, userValues []byte) (map[string]string, bool, error) {
	values, err := v3util.ReadValues(userValues)
	if err != nil {
		return nil, false, err
	}
	options := *t.releaseV3Option()

	// Check Release Name length
	if t.Release.Name != "" {
		err = v3util.ValidateReleaseName(t.Release.Name)
		if err != nil {
			return nil, false, err
		}
	}

	err = v3util.ProcessDependenciesWithMerge(targetChart, values)
	if err != nil {
		return nil, false, err
	}

	vals, err := v3util.ToRenderValuesWithSchemaValidation(targetChart, values.AsMap(), options, t.capabilitiesV3(), false)
	if err != nil {
		return nil, false, err
	}
	// When defaultTemplatesToAssert is empty, ensure all templates will be validated.
	if len(t.defaultTemplatesToAssert) == 0 {
		// Set all files
		t.defaultTemplatesToAssert = []string{multiWildcard}
	}

	// Filter the files that needs to be validated
	filteredChart := CopyV3Chart(t.chartRoute, targetChart.Name(), t.defaultTemplatesToAssert, targetChart)

	var outputOfFiles map[string]string
	// modify chart metadata before rendering
	t.ModifyChartMetadata(targetChart)
	if len(t.KubernetesProvider.Objects) > 0 {
		outputOfFiles, err = v3engine.RenderWithClientProvider(filteredChart, vals, &t.KubernetesProvider)
	} else {
		outputOfFiles, err = v3engine.Render(filteredChart, vals)
	}

	var renderSucceed bool
	outputOfFiles, renderSucceed, err = t.translateErrorToOutputFiles(err, outputOfFiles)
	log.WithField(LOG_TEST_JOB, "render-v3-chart").Debug("outputOfFiles:", outputOfFiles, "renderSucceed:", renderSucceed, "err:", err)
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

// capabilitiesV3 chartutil.Capabilities ready for render
// function returns a v3util.Capabilities struct based on the TestJob's capabilities.
// It overrides the KubeVersion field if majorVersion or minorVersion are set
func (t *TestJob) capabilitiesV3() *v3util.Capabilities {
	capabilities := v3util.DefaultCapabilities

	majorVersion := cmp.Or(t.Capabilities.MajorVersion, capabilities.KubeVersion.Major)
	minorVersion := cmp.Or(t.Capabilities.MinorVersion, capabilities.KubeVersion.Minor)

	capabilities.KubeVersion = v3util.KubeVersion{
		Version: fmt.Sprintf("v%s.%s.0", majorVersion, minorVersion),
		Major:   majorVersion,
		Minor:   minorVersion,
	}

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
		if assertion == nil {
			continue
		}
		result := assertion.Assert(
			manifestsOfFiles,
			snapshotComparer,
			renderSucceed,
			renderError,
			&results.AssertionResult{Index: idx},
			failfast,
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
		if assertion != nil {
			t.requireRenderSuccess = t.requireRenderSuccess && assertion.requireRenderSuccess
		}
	}
}

// add prefix to Assertion.Template
func (t *TestJob) polishAssertionsTemplate(targetChartName string, outputOfFiles map[string]string) {
	if t.chartRoute == "" {
		t.chartRoute = targetChartName
	}

	for _, assertion := range t.Assertions {
		if assertion == nil {
			continue
		}
		prefixedChartsNameFiles := false
		templatesToAssert := make([]string, 0)

		if t.DocumentIndex != nil {
			assertion.DocumentIndex = *t.DocumentIndex
		}

		if assertion.DocumentSelector == nil {
			assertion.DocumentSelector = t.DocumentSelector
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

// ModifyChartMetadata overrides the metadata of a Helm chart based on the values
// provided in the TestJob. If a chart version is set in the TestJob (t.Chart.Version),
// it updates the target chart's version and also propagates the same version to all
// chart dependencies. Similarly, if an appVersion is set (t.Chart.AppVersion),
// it updates the target chart's appVersion and also propagates it to all dependencies.
func (t *TestJob) ModifyChartMetadata(targetChart *v3chart.Chart) {
	targetChart.Metadata.Version = cmp.Or(t.Chart.Version, targetChart.Metadata.Version)
	targetChart.Metadata.AppVersion = cmp.Or(t.Chart.AppVersion, targetChart.Metadata.AppVersion)

	updateMetadata := func(version, appVersion string) {
		for _, dependency := range targetChart.Dependencies() {
			dependency.Metadata.Version = cmp.Or(version, dependency.Metadata.Version)
			dependency.Metadata.AppVersion = cmp.Or(appVersion, dependency.Metadata.AppVersion)
		}
	}
	updateMetadata(t.Chart.Version, t.Chart.AppVersion)
}

// SetCapabilities populates the Capabilities struct with values from CapabilitiesFields.
// It extracts majorVersion, minorVersion, and apiVersions fields and sets the corresponding
// fields in Capabilities. If apiVersions is nil, it sets APIVersions to nil. If it's a slice,
// it appends string values to APIVersions.
func (t *TestJob) SetCapabilities() {
	if val, ok := t.CapabilitiesFields["majorVersion"]; ok {
		t.Capabilities.MajorVersion = convertIToString(val)
	}
	if val, ok := t.CapabilitiesFields["minorVersion"]; ok {
		t.Capabilities.MinorVersion = convertIToString(val)
	}
	if val, ok := t.CapabilitiesFields["apiVersions"]; ok {
		switch v := val.(type) {
		case []interface{}:
			t.Capabilities.APIVersions = make([]string, 0, len(v)) // optimize slice allocation
			for _, item := range v {
				if str, ok := item.(string); ok {
					t.Capabilities.APIVersions = append(t.Capabilities.APIVersions, str)
				}
			}
		case nil:
		default:
			// key capabilities.apiVersions exists but is unset
			t.Capabilities.APIVersions = nil
		}
	} else {
		// APIVersions not set on test level
		t.Capabilities.APIVersions = make([]string, 0)
	}
}

// ConvertIToString The convertToString function takes an interface{} value as input and returns a string representation of it.
// If the input value is nil, it returns an empty string.
func convertIToString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	default:
		return ""
	}
}
