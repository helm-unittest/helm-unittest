package helmtest

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

type K8sManifest map[string]interface{}

type TestJob struct {
	defaultFile string
	Name        string `yaml:"it"`
	Values      []string
	Set         map[string]interface{}
	Assertions  []Assertion `yaml:"asserts"`
}

func (t TestJob) Run(targetChart *chart.Chart, result *TestJobResult) *TestJobResult {
	result.DisplayName = t.Name

	vv, err := t.vals()
	if err != nil {
		result.ExecError = err
		return result
	}

	config := &chart.Config{Raw: string(vv), Values: map[string]*chart.Value{}}
	options := chartutil.ReleaseOptions{
		Name:      "RELEASE_NAME",
		Time:      timeconv.Now(),
		Namespace: "NAMESPACE",
		//Revision:  1,
		//IsInstall: true,
	}

	// Set up engine.
	renderer := engine.New()

	vals, err := chartutil.ToRenderValues(targetChart, config, options)
	if err != nil {
		result.ExecError = err
		return result
	}

	outputOfFiles, err := renderer.Render(targetChart, vals)
	if err != nil {
		result.ExecError = err
		return result
	}

	manifestsOfFiles := make(map[string][]K8sManifest)
	for file, rendered := range outputOfFiles {
		documents := strings.Split(rendered, "---")
		manifests := make([]K8sManifest, len(documents))
		for i, doc := range documents {
			manifest := make(K8sManifest)
			if err := yaml.Unmarshal([]byte(doc), manifest); err != nil {
				result.ExecError = err
				return result
			}
			manifests[i] = manifest
		}
		manifestsOfFiles[path.Base(file)] = manifests
	}

	testPass := true
	assertsResult := make([]*AssertionResult, len(t.Assertions))
	for idx, assertion := range t.Assertions {
		if assertion.File == "" {
			assertion.File = t.defaultFile
		}
		result := assertion.Assert(manifestsOfFiles, &AssertionResult{Index: idx})
		assertsResult[idx] = result
		testPass = testPass && result.Passed
	}

	result.Passed = testPass
	result.AssertsResult = assertsResult
	return result
}

// liberally borrows from helm-template
func (t TestJob) vals() ([]byte, error) {
	base := map[interface{}]interface{}{}

	for _, valueFile := range t.Values {
		currentMap := map[interface{}]interface{}{}
		bytes, err := ioutil.ReadFile(valueFile)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", valueFile, err)
		}

		base = mergeValues(base, currentMap)
	}

	for path, valus := range t.Set {
		setMap, err := BuildValueOfSetPath(valus, path)
		if err != nil {
			return []byte{}, err
		}

		base = mergeValues(base, setMap)
	}
	return yaml.Marshal(base)
}
