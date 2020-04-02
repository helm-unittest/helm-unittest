package unittest

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
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
	v2engine "k8s.io/helm/pkg/engine"
	v2chart "k8s.io/helm/pkg/proto/hapi/chart"
	v2timeconv "k8s.io/helm/pkg/timeconv"
)

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
	Name       string `yaml:"it"`
	Values     []string
	Set        map[string]interface{}
	Assertions []*Assertion `yaml:"asserts"`
	Release    struct {
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
	// template assertion should assert if not specified
	defaultTemplateToAssert string
}

// RunV2 render the chart and validate it with assertions in TestJob.
func (t *TestJob) RunV2(
	targetChart *v2chart.Chart,
	cache *snapshot.Cache,
	result *TestJobResult,
) *TestJobResult {
	startTestRun := time.Now()
	t.polishV2AssertionsTemplate(targetChart)
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
	t.polishV3AssertionsTemplate(targetChart)
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

	for path, valus := range t.Set {
		setMap, err := valueutils.BuildValueOfSetPath(valus, path)
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
	options := *t.releaseV2Option()

	vals, err := v2util.ToRenderValues(targetChart, config, options)
	if err != nil {
		return nil, err
	}

	renderer := v2engine.New()
	outputOfFiles, err := renderer.Render(targetChart, vals)
	if err != nil {
		return nil, err
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

	vals, err := v3util.ToRenderValues(targetChart, values.AsMap(), options, v3util.DefaultCapabilities)
	if err != nil {
		return nil, err
	}

	outputOfFiles, err := v3engine.Render(targetChart, vals)
	if err != nil {
		return nil, err
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
		decoder := yaml.NewDecoder(strings.NewReader(rendered))

		if filepath.Ext(file) == ".yaml" {
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

			manifestsOfFiles[file] = manifests
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
func (t *TestJob) polishV2AssertionsTemplate(targetChart *v2chart.Chart) {
	if t.chartRoute == "" {
		t.chartRoute = targetChart.Metadata.Name
	}

	for _, assertion := range t.Assertions {
		var templateToAssert string

		if assertion.Template == "" {
			if t.defaultTemplateToAssert == "" {
				return
			}
			templateToAssert = t.defaultTemplateToAssert
		} else {
			templateToAssert = assertion.Template
		}

		// map the file name to the path of helm rendered result
		assertion.Template = filepath.ToSlash(
			filepath.Join(t.chartRoute, "templates", templateToAssert),
		)
	}
}

// add prefix to Assertion.Template
func (t *TestJob) polishV3AssertionsTemplate(targetChart *v3chart.Chart) {
	if t.chartRoute == "" {
		t.chartRoute = targetChart.Metadata.Name
	}

	for _, assertion := range t.Assertions {
		var templateToAssert string

		if assertion.Template == "" {
			if t.defaultTemplateToAssert == "" {
				return
			}
			templateToAssert = t.defaultTemplateToAssert
		} else {
			templateToAssert = assertion.Template
		}

		// map the file name to the path of helm rendered result
		assertion.Template = filepath.ToSlash(
			filepath.Join(t.chartRoute, "templates", templateToAssert),
		)
	}
}
