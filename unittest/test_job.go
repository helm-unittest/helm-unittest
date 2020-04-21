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

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/lrills/helm-unittest/unittest/validators"
	"github.com/lrills/helm-unittest/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"

	v3chart "helm.sh/helm/v3/pkg/chart"
	v3util "helm.sh/helm/v3/pkg/chartutil"
	v3engine "helm.sh/helm/v3/pkg/engine"
	v2util "k8s.io/helm/pkg/chartutil"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
	v2renderutil "k8s.io/helm/pkg/renderutil"
	v2timeconv "k8s.io/helm/pkg/timeconv"
)

func parseV2RenderError(errorMessage string) (string, string) {
	// Split the error into several groups.
	// those groups are required to parse the correct value.
	const regexPattern string = "^.+\"(.+)\":(.+:)* (.+)$"
	filePath := ""
	content := "<no value>"

	r := regexp.MustCompile(regexPattern)
	result := r.FindStringSubmatch(errorMessage)

	if len(result) == 4 {
		filePath = result[1]
		content = fmt.Sprintf("%s: %s", common.RAW, result[3])
	}

	return filePath, content
}

func parseV3RenderError(errorMessage string) (string, string) {
	// Split the error into several groups.
	// those groups are required to parse the correct value.
	const regexPattern string = "^.+\\((.+):\\d+:\\d+\\):(.+:)* (.+)$"
	filePath := ""
	content := "<no value>"

	r := regexp.MustCompile(regexPattern)
	result := r.FindStringSubmatch(errorMessage)

	if len(result) == 4 {
		filePath = result[1]
		content = fmt.Sprintf("%s: %s", common.RAW, result[3])
	}

	return filePath, content
}

func parseYamlFile(rendered string) ([]common.K8sManifest, error) {
	decoder := yaml.NewDecoder(strings.NewReader(rendered))
	manifests := make([]common.K8sManifest, 0)

	for {
		manifest := make(common.K8sManifest)
		if err := decoder.Decode(manifest); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		if len(manifest) > 0 {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
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
	DocumentIndex *int         `yaml:"documentIndex"`
	Assertions    []*Assertion `yaml:"asserts"`
	Release       struct {
		Name      string
		Namespace string
		Revision  int
		IsUpgrade bool
	}
	// route indicate which chart in the dependency hierarchy
	// like "parant-chart", "parent-charts/charts/child-chart"
	chartRoute string
	// where the test suite file located
	definitionFile string
	// list of templates assertion should assert if not specified
	defaultTemplatesToAssert []string
}

// RunV2 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV2(
	targetChart *v2chart.Chart,
	cache *snapshot.Cache,
	result *TestJobResult,
) *TestJobResult {
	startTestRun := time.Now()
	t.polishAssertionsTemplate(targetChart.Metadata.Name)
	result.DisplayName = t.Name

	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, err := t.renderV2Chart(targetChart, userValues)
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
	)

	result.Duration = time.Now().Sub(startTestRun)
	return result
}

// RunV3 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV3(
	targetChart *v3chart.Chart,
	cache *snapshot.Cache,
	result *TestJobResult,
) *TestJobResult {
	startTestRun := time.Now()
	t.polishAssertionsTemplate(targetChart.Name())
	result.DisplayName = t.Name

	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, err := t.renderV3Chart(targetChart, userValues)
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
	)

	result.Duration = time.Now().Sub(startTestRun)
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

// render the chart and return result map
func (t *TestJob) renderV2Chart(targetChart *v2chart.Chart, userValues []byte) (map[string]string, error) {
	config := &v2chart.Config{Raw: string(userValues)}
	defaultKubeVersion := fmt.Sprintf("%s.%s", v2util.DefaultKubeVersion.Major, v2util.DefaultKubeVersion.Minor)
	renderOpts := v2renderutil.Options{
		ReleaseOptions: *t.releaseV2Option(),
		KubeVersion:    defaultKubeVersion,
		APIVersions:    []string{},
	}

	outputOfFiles, err := v2renderutil.Render(targetChart, config, renderOpts)
	// When rendering failed, due to fail or required,
	// make sure to translate the error to outputOfFiles.
	if err != nil {
		// Parse the error and create an outputFile
		filePath, content := parseV2RenderError(err.Error())
		// If error not parsed well, rethrow as normal.
		if filePath == "" {
			return nil, err
		}
		outputOfFiles[filePath] = content
	}

	return outputOfFiles, nil
}

// render the chart and return result map
func (t *TestJob) renderV3Chart(targetChart *v3chart.Chart, userValues []byte) (map[string]string, error) {
	values, err := v3util.ReadValues(userValues)
	if err != nil {
		return nil, err
	}
	options := *t.releaseV3Option()

	err = v3util.ProcessDependencies(targetChart, values)
	if err != nil {
		return nil, err
	}

	vals, err := v3util.ToRenderValues(targetChart, values.AsMap(), options, v3util.DefaultCapabilities)
	if err != nil {
		return nil, err
	}

	outputOfFiles, err := v3engine.Render(targetChart, vals)
	// When rendering failed, due to fail or required,
	// make sure to translate the error to outputOfFiles.
	if err != nil {
		// Parse the error and create an outputFile
		filePath, content := parseV3RenderError(err.Error())
		// If error not parsed well, rethrow as normal.
		if filePath == "" {
			return nil, err
		}
		outputOfFiles[filePath] = content
	}

	return outputOfFiles, nil
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
) (bool, []*AssertionResult) {
	testPass := true
	assertsResult := make([]*AssertionResult, len(t.Assertions))

	for idx, assertion := range t.Assertions {
		result := assertion.Assert(
			manifestsOfFiles,
			snapshotComparer,
			&AssertionResult{Index: idx},
		)

		assertsResult[idx] = result
		testPass = testPass && result.Passed
	}
	return testPass, assertsResult
}

// add prefix to Assertion.Template
func (t *TestJob) polishAssertionsTemplate(targetChartName string) {
	if t.chartRoute == "" {
		t.chartRoute = targetChartName
	}

	for _, assertion := range t.Assertions {
		templatesToAssert := make([]string, 0)

		if t.DocumentIndex != nil {
			assertion.DocumentIndex = *t.DocumentIndex
		}

		if assertion.Template == "" {
			if t.Template == "" {
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
