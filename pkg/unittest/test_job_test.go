package unittest_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/stretchr/testify/mock"

	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	a := assert.New(t)
	a.Equal(tj.Name, "should do something")
	a.Equal(tj.Values, []string{"values.yaml"})
	a.Equal(tj.Set, map[string]interface{}{
		"a.b.c": "ABC",
		"x.y.z": "XYZ",
	})
	assertions := make([]*Assertion, 2)
	assErr := common.YmlUnmarshal(`
  - equal:
      path: a.b
      value: c
  - matchRegex:
      path: x.y
      pattern: /z/
`, &assertions)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
		WithRenderPath(renderPath),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})
	defer os.RemoveAll(renderPath)

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)
	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{}))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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

	file := path.Join("_scratch", "testjob_test_TestRunJobWithValuesFile.yaml")
	a.Nil(writeToFile("nameOverride: mary-jane", file))
	defer os.RemoveAll(file)

	var tj TestJob
	common.YmlUnmarshalTestHelper(fmt.Sprintf(manifest, file), &tj, t)
	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.FileExists(file)
	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)
	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cfg := NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	)
	tj.WithConfig(*cfg)
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)
	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})
	assert.False(t, testResult.Passed)
	assert.Equal(t, "7", tj.Capabilities.MinorVersion)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)
	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a := assert.New(t)
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)
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
	unmarshalJobTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

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
	unmarshalJobTestHelper(manifest, &tj, t)
	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	assert.Nil(t, testResult.ExecError)
	assert.True(t, testResult.Passed)
	assert.Equal(t, 1, len(testResult.AssertsResult))
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
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

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
	common.YmlUnmarshal(manifest, &tj)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	a := assert.New(t)

	a.Nil(testResult.ExecError)
	a.Equal(1, len(testResult.AssertsResult))
	a.Equal(testResult.AssertsResult[0].AssertType, "equal")
	a.NotEqual(testResult.AssertsResult[0].AssertType, "notSupportedAssert")
}

func TestV3RunJobPatchArrayValue(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should patch second element
template: templates/configmap.yaml
values:
- %s
set:
 ingress:
  hosts[1]: should-override-second.local
asserts:
  - exists:
      path: data["my.ingress.hosts"]
  - notEqual:
      path: data["my.ingress.hosts"]
      value:
      - null
      - should-override-second.local
  - equal:
      path: data["my.ingress.hosts"]
      value:
      - chart-example-first.local
      - chart-example-second.local
      - chart-example-third.local
`
	a := assert.New(t)
	file := path.Join("_scratch", "test_tmp_values.yaml")
	a.Nil(writeToFile(`
ingress:
  hosts:
    - chart-example-first.local
    - chart-example-second.local
    - chart-example-third.local
`, file))
	defer os.RemoveAll(file)

	var tj TestJob
	unmarshalJobTestHelper(fmt.Sprintf(manifest, file), &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	assert.Nil(t, testResult.ExecError)
	assert.True(t, testResult.Passed)
	assert.Equal(t, 3, len(testResult.AssertsResult))
}

func TestV3RunJobPatchPointArrayValue(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should patch second element
template: templates/configmap.yaml
values:
- %s
set:
 ingress.hosts[1]: should-override-second.local
asserts:
  - exists:
      path: data["my.ingress.hosts"]
  - equal:
      path: data["my.ingress.hosts"]
      value:
      - null
      - should-override-second.local
  - notEqual:
      path: data["my.ingress.hosts"]
      value:
      - chart-example-first.local
      - should-override-second.local
      - chart-example-third.local
`
	a := assert.New(t)
	file := path.Join("_scratch", "test_tmp_values.yaml")
	a.Nil(writeToFile(`
ingress:
  hosts:
    - chart-example-first.local
    - chart-example-second.local
    - chart-example-third.local
`, file))
	defer os.RemoveAll(file)

	var tj TestJob
	unmarshalJobTestHelper(fmt.Sprintf(manifest, file), &tj, t)

	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
		WithFailFast(true),
	))
	testResult := tj.RunV3(&results.TestJobResult{})

	assert.Nil(t, testResult.ExecError)
	assert.True(t, testResult.Passed)
	assert.Equal(t, 3, len(testResult.AssertsResult))
}

func TestV3RunJob_TplFunction_Fail_WithoutAssertion(t *testing.T) {
	c := &v3chart.Chart{
		Metadata: &v3chart.Metadata{
			Name:    "moby",
			Version: "1.2.3",
		},
		Templates: []*v3chart.File{},
		Values:    map[string]interface{}{},
	}

	tests := []struct {
		template *v3chart.File
		error    error
	}{
		{
			template: &v3chart.File{Name: "templates/validate.tpl", Data: []byte("{{- fail (printf \"`root`\") }}")},
			error:    errors.New("execution error at (moby/templates/validate.tpl:1:4): `root`"),
		},
		{
			template: &v3chart.File{Name: "templates/validate.tpl", Data: []byte("{{- fail (printf \"\n`root`\") }}")},
			error:    errors.New("parse error at (moby/templates/validate.tpl:1): unterminated quoted string"),
		},
	}

	a := assert.New(t)

	for _, test := range tests {
		tj := TestJob{}
		c.Templates = []*v3chart.File{test.template}
		tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
			WithFailFast(true),
		))
		testResult := tj.RunV3(&results.TestJobResult{})
		a.Error(testResult.ExecError)
		a.False(testResult.Passed)
		a.EqualError(testResult.ExecError, test.error.Error())
	}
}

func TestV3RunJob_TplFunction_Fail_WithAssertion(t *testing.T) {
	c := &v3chart.Chart{
		Metadata: &v3chart.Metadata{
			Name:    "moby",
			Version: "1.2.3",
		},
		Templates: []*v3chart.File{},
		Values:    map[string]interface{}{},
	}

	tests := []struct {
		template *v3chart.File
		error    error
		expected bool
	}{
		{
			template: &v3chart.File{Name: "templates/validate.tpl", Data: []byte("{{- fail (printf \"`root`\") }}")},
			error:    nil,
			expected: true,
		},
		{
			template: &v3chart.File{Name: "templates/validate.tpl", Data: []byte("{{- fail (printf \"\n`root`\") }}")},
			error:    errors.New("parse error at (moby/templates/validate.tpl:1): unterminated quoted string"),
			expected: false,
		},
	}

	manifest := `
it: should validate failure message
template: templates/validate.tpl
asserts:
- failedTemplate:
    errorPattern: "` + "`root`" + `"
`
	var tj TestJob
	unmarshalJobTestHelper(manifest, &tj, t)

	for _, test := range tests {
		c.Templates = []*v3chart.File{test.template}
		tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{},
			WithFailFast(true),
		))
		testResult := tj.RunV3(&results.TestJobResult{})
		assert.Equal(t, test.expected, testResult.Passed)
		if test.error != nil {
			assert.NotNil(t, testResult.ExecError)
			assert.False(t, testResult.Passed)
			assert.EqualError(t, testResult.ExecError, test.error.Error())
		} else {
			assert.Nil(t, testResult.ExecError)
			assert.True(t, testResult.Passed)
		}
	}
}

const fileKeyPrefix = "#### file:"

func Test_SplitManifests(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedManifests map[string]string
	}{
		{
			name:  "Single Manifest",
			input: "---\n" + fileKeyPrefix + " test-key\nmanifest1\n", // Literal separator
			expectedManifests: map[string]string{
				"test-key": "manifest1\n",
			},
		},
		{
			name:  "Multiple Manifests",
			input: "---\n" + fileKeyPrefix + " test-key1\nmanifest1\n" + "---\n" + fileKeyPrefix + " test-key2\nmanifest2\n",
			expectedManifests: map[string]string{
				"test-key1": "manifest1\n",
				"test-key2": "manifest2\n",
			},
		},
		{
			name:  "Multi-document Manifest",
			input: "---\n" + fileKeyPrefix + " test-key\nmanifest1\n---\nmanifest2\n",
			expectedManifests: map[string]string{
				"test-key": "manifest1\n---\nmanifest2\n",
			},
		},
		{
			// TODO: should we treat the post-renderer handing us an empty file as "a map of one empty file?"
			// or should we return an empty map?
			name:  "Empty Input",
			input: "",
			expectedManifests: map[string]string{
				"manifest.yaml": "",
			},
		},
		{
			name:  "Manifest with no newline",
			input: "---\n" + fileKeyPrefix + " test-key\nmanifest1",
			expectedManifests: map[string]string{
				"test-key": "manifest1",
			},
		},
		{
			name:  "Manifest with multiple newlines",
			input: "---\n" + fileKeyPrefix + " test-key\nmanifest1\n\n",
			expectedManifests: map[string]string{
				"test-key": "manifest1\n\n",
			},
		},
		{
			name:  "Manifest with net-new content",
			input: "---\nmanifest1\n",
			expectedManifests: map[string]string{
				"manifest.yaml": "---\nmanifest1\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputBuffer := bytes.NewBufferString(test.input)
			actualManifests := SplitManifests(inputBuffer)
			assert.Equal(t, test.expectedManifests, actualManifests)
		})
	}
}

func Test_MergeAndPostRender(t *testing.T) {
	tests := []struct {
		name           string
		inputManifests map[string]string
		expectedOutput string
		postRenderer   *mockPostRenderer // Use the mock type
		expectedInput  string            // Add expected input for postRenderer
	}{
		{
			name: "With Post-render newlines",
			inputManifests: map[string]string{
				"test-key1": "manifest1\n",
				"test-key2": "manifest2\n",
			},
			expectedOutput: "---\n" + fileKeyPrefix + " test-key1\nmanifest1-modified\n\n---\n" + fileKeyPrefix + " test-key2\nmanifest2-modified\n\n",
			postRenderer:   &mockPostRenderer{},
			expectedInput:  "---\n" + fileKeyPrefix + " test-key1\nmanifest1\n---\n" + fileKeyPrefix + " test-key2\nmanifest2\n", // Input *before* post-rendering
		},
		{
			name: "With Post-render no newlines",
			inputManifests: map[string]string{
				"test-key1": "manifest1\n",
				"test-key2": "manifest2\n",
			},
			expectedOutput: "---\n" + fileKeyPrefix + " test-key1\nmanifest1-modified\n---\n" + fileKeyPrefix + " test-key2\nmanifest2-modified",
			postRenderer:   &mockPostRenderer{},
			expectedInput:  "---\n" + fileKeyPrefix + " test-key1\nmanifest1\n---\n" + fileKeyPrefix + " test-key2\nmanifest2\n", // Input *before* post-rendering
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up the mock expectation *before* calling MergeAndPostRender
			expectedBuffer := bytes.NewBufferString(test.expectedInput)
			test.postRenderer.On("Run", mock.Anything).Return(bytes.NewBufferString(test.expectedOutput), nil).Run(func(args mock.Arguments) {
				arg := args.Get(0).(*bytes.Buffer)
				assert.Equal(t, expectedBuffer.String(), arg.String())
			})

			output, err := MergeAndPostRender(test.inputManifests, test.postRenderer)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedOutput, output.String())

			test.postRenderer.AssertExpectations(t) // Verify the mock was called as expected

		})
	}
}

type mockPostRenderer struct {
	mock.Mock
	output string
	err    error
}

func (m *mockPostRenderer) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	args := m.Called(renderedManifests)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

// func TestV3RunJobWithSuccessWhenNoDocumentSelectorSkipEmptyTemplateAndNoTemplatesUnderTest(t *testing.T) {
// 	c, _ := loader.Load(testV3BasicChart)
// 	manifest := `
// it: should work
// set:
//   image.tag: ""
// chart:
//   version: 9.9.9+test
//   appVersion: 9999
// template: templates/deployment.yaml
// documentSelector:
//   path: kind
//   value: SomeKind
//   skipEmptyTemplate: true
// asserts:
//   - equal:
//       path: metadata.labels.chart
//       value: basic-9.9.9_test
//   - equal:
//       path: spec.template.spec.containers[0].image
//       value: nginx:9999
// `
// 	var tj TestJob
// 	unmarshalJobTestHelper(manifest, &tj, t)
//
// 	tj.WithConfig(*NewTestConfig(c, &snapshot.Cache{}, WithEmtpyTemplatesSkipped(true)))
// 	testResult := tj.RunV3(&results.TestJobResult{})
//
// 	a := assert.New(t)
// 	a.Nil(testResult.ExecError)
// 	a.True(testResult.Passed)
// 	fmt.Println(testResult.Passed)
// 	a.Equal(2, len(testResult.AssertsResult))
// }
