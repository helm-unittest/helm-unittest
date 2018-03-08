package unittest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/lrills/helm-unittest/unittest/validators"
	"github.com/lrills/helm-unittest/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
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

// TestJob defintion of a test, including values and assertions
type TestJob struct {
	Name       string `yaml:"it"`
	Values     []string
	Set        map[string]interface{}
	Assertions []*Assertion `yaml:"expects"`
	Release    struct {
		Name      string
		Namespace string
		Revision  int
		IsUpgrade bool
	}
	definitionFile          string
	defaultTemplateToAssert string
}

// Run render the chart and validate it with assertions in TestJob
func (t *TestJob) Run(
	targetChart *chart.Chart,
	cache *snapshot.Cache,
	result *TestJobResult,
) *TestJobResult {
	result.DisplayName = t.Name

	userValues, err := t.getUserValues()
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, err := t.renderChart(targetChart, userValues)
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

	return result
}

// liberally borrows from helm-template
func (t *TestJob) getUserValues() ([]byte, error) {
	base := map[interface{}]interface{}{}

	for _, valueFile := range t.Values {
		currentMap := map[interface{}]interface{}{}
		valueFilePath := filepath.Join(filepath.Dir(t.definitionFile), valueFile)
		bytes, err := ioutil.ReadFile(valueFilePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", valueFile, err)
		}
		base = valueutils.MergeValues(base, currentMap)
	}

	for path, valus := range t.Set {
		setMap, err := valueutils.BuildValueOfSetPath(valus, path)
		if err != nil {
			return []byte{}, err
		}

		base = valueutils.MergeValues(base, setMap)
	}
	return yaml.Marshal(base)
}

func (t *TestJob) renderChart(targetChart *chart.Chart, userValues []byte) (map[string]string, error) {
	config := &chart.Config{Raw: string(userValues), Values: map[string]*chart.Value{}}
	options := *t.releaseOption()

	vals, err := chartutil.ToRenderValues(targetChart, config, options)
	if err != nil {
		return nil, err
	}

	renderer := engine.New()
	outputOfFiles, err := renderer.Render(targetChart, vals)
	if err != nil {
		return nil, err
	}
	return outputOfFiles, nil
}

func (t *TestJob) releaseOption() *chartutil.ReleaseOptions {
	options := chartutil.ReleaseOptions{
		Name:      "RELEASE-NAME",
		Namespace: "NAMESPACE",
		Time:      timeconv.Now(),
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

func (t *TestJob) parseManifestsFromOutputOfFiles(outputOfFiles map[string]string) (
	map[string][]common.K8sManifest,
	error,
) {
	manifestsOfFiles := make(map[string][]common.K8sManifest)
	for file, rendered := range outputOfFiles {
		documents := strings.Split(rendered, "---")
		manifests := make([]common.K8sManifest, len(documents))

		manifestCount := 0
		for _, doc := range documents {
			manifest := make(common.K8sManifest)
			if err := yaml.Unmarshal([]byte(doc), manifest); err != nil {
				return nil, err
			}

			if len(manifest) > 0 {
				manifests[manifestCount] = manifest
				manifestCount++
			}
		}
		manifestsOfFiles[filepath.Base(file)] = manifests[:manifestCount]
	}
	return manifestsOfFiles, nil
}

func (t *TestJob) runAssertions(
	manifestsOfFiles map[string][]common.K8sManifest,
	comparer validators.SnapshotComparer,
) (bool, []*AssertionResult) {
	testPass := true
	assertsResult := make([]*AssertionResult, len(t.Assertions))

	for idx, assertion := range t.Assertions {
		result := assertion.Assert(
			manifestsOfFiles,
			comparer,
			&AssertionResult{Index: idx},
		)

		assertsResult[idx] = result
		testPass = testPass && result.Passed
	}
	return testPass, assertsResult
}

func (t *TestJob) polishAssertions() {
	for _, assertion := range t.Assertions {
		if assertion.Template == "" {
			assertion.Template = t.defaultTemplateToAssert
		}
	}
}
