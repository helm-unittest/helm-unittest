package unittest_test

import (
	"fmt"
	"os"

	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	v3chart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func makeTestJobResultSnapshotable(result *results.TestJobResult) *results.TestJobResult {
	result.Duration, _ = time.ParseDuration("0s")
	return result
}

func TestUnmarshalableJobFromYAML(t *testing.T) {
	manifest := `
it: should do something
values:
  - values.yaml
set:
  a.b.c: ABC
  x.y.z: XYZ
asserts:
  - equal:
      path: a.b
      value: c
  - matchRegex:
      path: x.y
      pattern: /z/
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	a := assert.New(t)
	a.Nil(err)
	a.Equal(tj.Name, "should do something")
	a.Equal(tj.Values, []string{"values.yaml"})
	a.Equal(tj.Set, map[string]interface{}{
		"a.b.c": "ABC",
		"x.y.z": "XYZ",
	})
	assertions := make([]*Assertion, 2)
	assErr := yaml.Unmarshal([]byte(`
  - equal:
      path: a.b
      value: c
  - matchRegex:
      path: x.y
      pattern: /z/
`), &assertions)
	a.Nil(assErr)
	a.Equal(tj.Assertions, assertions)
}

func TestV3RunJobOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    template: templates/deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    documentSelector:
      path: metadata.name
      value: RELEASE-NAME-basic
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithRenderPathOk(t *testing.T) {
	renderPath := "testdebug"
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    template: templates/deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    documentSelector:
      path: metadata.name
      value: RELEASE-NAME-basic
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, renderPath, &results.TestJobResult{})
	defer os.RemoveAll(renderPath)

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
	// Check folder contains files
	a.DirExists(renderPath)
}

func TestV3RunJobWithTestJobTemplateOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
template: templates/deployment.yaml
documentIndex: 0
asserts:
  - equal:
      path: kind
      value: Deployment
  - matchRegex:
      path: metadata.name
      pattern: -basic$
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithTestJobDocumentSelectorOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
template: templates/deployment.yaml
documentSelector:
  path: metadata.name
  value: RELEASE-NAME-basic-db
asserts:
  - equal:
      path: kind
      value: Deployment
  - matchRegex:
      path: metadata.name
      pattern: -basic
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithTestJobTemplatesOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
templates:
  - templates/deployment.yaml
  - templates/configmap.yaml
asserts:
  - equal:
      path: kind
      value: Deployment
    template: templates/deployment.yaml
  - equal:
      path: kind
      value: ConfigMap
    template: templates/configmap.yaml
  - exists:
      path: metadata.name
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(3, len(testResult.AssertsResult))
}

func TestV3RunJobWithTestMissingRequiredValueOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
set:
  ingress.enabled: true
  service.externalPort: ""
template: templates/ingress.yaml
asserts:
  - failedTemplate:
      errorMessage: The externalPort is required
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithAssertionFail(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    documentIndex: 0
    template: templates/deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, false, "", &results.TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithAssertionFailFast(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    documentIndex: 0
    template: templates/deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithValueSet(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
set:
  nameOverride: john-doe
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-john-doe
    documentIndex: 0
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithValuesFile(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
values:
  - %s
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-mary-jane
    documentIndex: 0
    template: templates/deployment.yaml
`
	a := assert.New(t)

	file, fileErr := os.Create("testjob_test_TestRunJobWithValuesFile.yaml")
	if fileErr != nil {
		a.FailNow("Failed to create file")
	}
	_, writeErr := file.WriteString("nameOverride: mary-jane")
	if writeErr != nil {
		a.FailNow("Failed to write to file")
	}
	closeErr := file.Close()
	if closeErr != nil {
		a.FailNow("Failed to close file")
	}

	var tj TestJob
	err := yaml.Unmarshal([]byte(fmt.Sprintf(manifest, file.Name())), &tj)
	a.Nil(err)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.FileExists(file.Name())
	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))

	os.Remove(file.Name())
}

func TestV3RunJobWithReleaseSettings(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
release:
  name: my-release
  namespace: test
asserts:
  - equal:
      path: metadata.name
      value: my-release-basic
    documentIndex: 0
    template: templates/deployment.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithNoCapabilitySettingsEmptyDoc(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - hasDocuments:
      count: 0
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithTooLongReleaseName(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: to long releasename
release:
  name: my-very-very-very-very-very-very-very-very-very-very-very-very-release
asserts:
  - hasDocuments:
      count: 1
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.NotNil(testResult.ExecError)
	a.False(testResult.Passed)
}

func TestV3RunJobWithCapabilitySettings(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
capabilities:
  majorVersion: 1
  minorVersion: 12
  apiVersions:
    - br.dev.local/v1
asserts:
  - hasDocuments:
      count: 1
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	a := assert.New(t)
	a.Nil(err)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithCapabilityApiVersionUnset(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
capabilities:
  majorVersion: 1
  minorVersion: 12
  apiVersions:
asserts:
  - hasDocuments:
      count: 1
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	a := assert.New(t)
	a.Nil(err)

	testResult := tj.RunV3(c, nil, true, "", &results.TestJobResult{})
	a.False(testResult.Passed)
	a.Equal(0, len(tj.Capabilities.APIVersions))
}

func TestV3RunJobWithCapabilityApiVersionNotSet(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
capabilities:
  majorVersion: 1
  minorVersion: 12
asserts:
  - hasDocuments:
      count: 1
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	a := assert.New(t)
	a.Nil(err)

	testResult := tj.RunV3(c, nil, true, "", &results.TestJobResult{})

	a.False(testResult.Passed)
	a.Equal([]string{}, tj.Capabilities.APIVersions)
	a.Equal("12", tj.Capabilities.MinorVersion)
}

func TestV3RunJobWithCapabilityMinorVersionSet(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
capabilities:
  minorVersion: 7
asserts:
  - hasDocuments:
      count: 1
    template: templates/crd_backup.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	a := assert.New(t)
	a.Nil(err)

	testResult := tj.RunV3(c, nil, true, "", &results.TestJobResult{})
	a.False(testResult.Passed)
	a.Equal("7", tj.Capabilities.MinorVersion)
}

func TestV3RunJobWithChartSettings(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
set:
  image.tag: ""
chart:
  version: 9.9.9+test
  appVersion: 9999
asserts:
  - equal:
      path: metadata.labels.chart
      value: basic-9.9.9_test
    template: templates/deployment.yaml
  - equal:
      path: spec.template.spec.containers[0].image
      value: nginx:9999
    template: templates/deployment.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithFailingTemplate(t *testing.T) {
	c, _ := loader.Load(testV3WithFailingTemplateChart)
	manifest := `
it: should load complete chart and validate configMap
release:
  name: ab
asserts:
  - failedTemplate:
      errorMessage:	"error calling include: template: no template \"non-existing-named-template\" associated with template \"gotpl\""
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithSchema(t *testing.T) {
	c, _ := loader.Load(testV3WithSchemaChart)
	manifest := `
it: should work
template: templates/dummy.yaml
asserts:
  - failedTemplate:
      errorMessage: "values don't meet the specifications of the schema(s) in the following chart(s):\nwith-schema:\n- (root): image is required\n"
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.NotNil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithSchemaAndNullValue(t *testing.T) {
	c, _ := loader.Load(testV3WithSchemaChart)
	manifest := `
it: should work
set:
  image:
    repository: "repo"
    pullPolicy: IfNotPresent
  value: null
asserts:
  - failedTemplate: {}
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.NotNil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithSchemaOk(t *testing.T) {
	c, _ := loader.Load(testV3WithSchemaChart)
	manifest := `
it: should work
template: templates/dummy.yaml
set:
  image:
    repository: "repo"
    pullPolicy: IfNotPresent
asserts:
  - notFailedTemplate: {}
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunSubChartWithVersionOverride(t *testing.T) {
	c, _ := loader.Load(testV3WithSubChart)
	manifest := `
it: should contain subchart and alias subchart when chart version is explicitly set
chart:
  version: 1.2.3
templates:
- charts/another-postgresql/templates/deployment.yaml
- charts/postgresql/templates/deployment.yaml
asserts:
  - matchRegex:
      path: metadata.labels["chart"]
      pattern: "(.*-)?postgresql-1.2.3"
`
	var tj TestJob
	a := assert.New(t)
	err := unmarshalJob(manifest, &tj)
	a.Nil(err)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunSubChartWithoutVersionOverride(t *testing.T) {
	c, _ := loader.Load(testV3WithSubChart)
	manifest := `
it: should contain subchart and alias subchart without version override
templates:
- charts/another-postgresql/templates/deployment.yaml
- charts/postgresql/templates/deployment.yaml
asserts:
  - matchRegex:
      path: metadata.labels["chart"]
      pattern: "(.*-)?postgresql-0.8.3"
`
	var tj TestJob
	a := assert.New(t)
	err := unmarshalJob(manifest, &tj)
	a.Nil(err)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestModifyChartMetadataVersions(t *testing.T) {
	type ChartVersions struct {
		Version    string
		AppVersion string
	}
	tests := []struct {
		name                 string
		testJob              TestJob
		initialChartVersions ChartVersions
		dependencies         []*v3chart.Chart // List of dependencies to add
		expectedVersions     ChartVersions
	}{
		{
			name: "Override version and propagate to dependencies",
			testJob: TestJob{
				Chart: struct {
					Version    string
					AppVersion string `yaml:"appVersion"`
				}{
					Version:    "1.2.3",
					AppVersion: "1.1.0",
				},
			},
			initialChartVersions: ChartVersions{"1.0.3", "1.0.8"},
			dependencies: []*v3chart.Chart{
				{Metadata: &v3chart.Metadata{Version: "0.1.0"}},
				{Metadata: &v3chart.Metadata{AppVersion: "11.1.0"}},
			},
			expectedVersions: ChartVersions{"1.2.3", "1.1.0"},
		},
		{
			name: "Override appVersion and propagate to dependencies",
			testJob: TestJob{
				Chart: struct {
					Version    string
					AppVersion string `yaml:"appVersion"`
				}{
					AppVersion: "2.0.0",
				},
			},
			initialChartVersions: ChartVersions{AppVersion: "1.0.0"},
			dependencies: []*v3chart.Chart{
				{Metadata: &v3chart.Metadata{AppVersion: "1.0.0"}},
			},
			expectedVersions: ChartVersions{AppVersion: "2.0.0"},
		},
		{
			name: "No overrides when TestJob has empty version and appVersion",
			testJob: TestJob{
				Chart: struct {
					Version    string
					AppVersion string `yaml:"appVersion"`
				}{},
			},
			initialChartVersions: ChartVersions{Version: "0.1.0", AppVersion: "1.0.0"},
			dependencies: []*v3chart.Chart{
				{Metadata: &v3chart.Metadata{Version: "0.1.0", AppVersion: "1.0.0"}},
				{Metadata: &v3chart.Metadata{Version: "0.1.0", AppVersion: "1.0.0"}},
			},
			expectedVersions: ChartVersions{Version: "0.1.0", AppVersion: "1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &v3chart.Chart{
				Metadata: &v3chart.Metadata{
					Version:    tt.expectedVersions.Version,
					AppVersion: tt.expectedVersions.AppVersion,
				},
			}

			// Add dependencies charts
			for _, dep := range tt.dependencies {
				parent.AddDependency(dep)
			}

			// Call the method to test
			tt.testJob.ModifyChartMetadata(parent)

			// Assert chart version
			assert.Equal(t, tt.expectedVersions.Version, parent.Metadata.Version)
			assert.Equal(t, tt.expectedVersions.AppVersion, parent.Metadata.AppVersion)

			// Assert dependencies' version and appVersion using the Dependencies() method
			for _, dep := range parent.Dependencies() {
				assert.Equal(t, tt.expectedVersions.Version, dep.Metadata.Version)
				assert.Equal(t, tt.expectedVersions.AppVersion, dep.Metadata.AppVersion)
			}
		})
	}
}

func TestV3RunJobWithTestJobNotesOk(t *testing.T) {
	c, _ := loader.Load(testV3WithSubChart)
	manifest := `
it: should generate notes
template: charts/child-chart/templates/NOTES-with-separator.txt
release:
  name: test-unit-notes
  namespace: unit-test
asserts:
 - equalRaw:
     value: |
       -----
       Platform release "test-unit-notes" installed in namespace "unit-test"

       Documentation can be found here: https://docs.example.com/
       -----
`
	var tj TestJob
	assert := assert.New(t)
	err := yaml.Unmarshal([]byte(manifest), &tj)
	assert.NoError(err)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	assert.NoError(testResult.ExecError)
	assert.True(testResult.Passed, testResult.AssertsResult)
}

func TestV3RunJobWithWithNotSupportedAssert(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should skip when not supported assert is found
template: templates/deployment.yaml
documentIndex: 0
asserts:
  - equal:
      path: kind
      value: Deployment
  - notSupportedAssert:
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, "", &results.TestJobResult{})

	a := assert.New(t)

	a.Nil(testResult.ExecError)
	a.Equal(1, len(testResult.AssertsResult))
	a.Equal(testResult.AssertsResult[0].AssertType, "equal")
	a.NotEqual(testResult.AssertsResult[0].AssertType, "notSupportedAssert")
}

func TestV3RunJobReplaceSlice(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
set:
 ingress:
  hosts[1]: chart-example.suite
asserts:
  - exists:
      path: spec.rules[?(@.host == "chart-example.suite")]
    template: templates/ingress.yaml
`
	var tj TestJob
	err := unmarshalJob(manifest, &tj)

	testResult := tj.RunV3(c, nil, true, "", &results.TestJobResult{})

	a := assert.New(t)
	a.Nil(err)

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	// a.Equal(2, len(testResult.AssertsResult))
}

// job with subchart
