package unittest_test

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/lrills/helm-unittest/pkg/unittest"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"

	v2util "k8s.io/helm/pkg/chartutil"
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
	yaml.Unmarshal([]byte(`
  - equal:
      path: a.b
      value: c
  - matchRegex:
      path: x.y
      pattern: /z/
`), &assertions)
	a.Equal(tj.Assertions, assertions)
}

func TestV2RunJobOk(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    documentIndex: 0
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithNOTESTemplateOk(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
template: NOTES.txt
asserts:
  - equalRaw:
      value: | 
        1. Get the application URL by running these commands:
          export POD_NAME=$(kubectl get pods --namespace NAMESPACE -l "app=basic,release=RELEASE-NAME" -o jsonpath="{.items[0].metadata.name}")
          echo "Visit http://127.0.0.1:8080 to use your application"
          kubectl port-forward $POD_NAME 8080:80
  - matchRegexRaw:
      pattern: ^1. Get the application URL by running these commands
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithTestJobTemplateOk(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
template: deployment.yaml
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
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithTestJobTemplatesOk(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
templates:
  - deployment.yaml
  - configmap.yaml
asserts:
  - equal:
      path: kind
      value: Deployment   
    template: deployment.yaml
  - equal:
      path: kind
      value: ConfigMap   
    template: configmap.yaml
  - isNotNull:
      path: metadata.name
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(3, len(testResult.AssertsResult))
}

func TestV2RunJobWithTestMissingRequiredValueOk(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
set:
  ingress.enabled: true
  service.externalPort: ""
template: ingress.yaml
asserts:
  - failedTemplate:
      errorMessage: The externalPort is required
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithAssertionFail(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    documentIndex: 0
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, false, &results.TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithAssertionFailFast(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    documentIndex: 0
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithValueSet(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
set:
  nameOverride: john-doe
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-john-doe
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithValuesFile(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
values:
  - %s
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-mary-jane
    documentIndex: 0
    template: deployment.yaml
`
	file, _ := ioutil.TempFile("", "testjob_test_TestRunJobWithValuesFile.yaml")
	file.WriteString("nameOverride: mary-jane")

	var tj TestJob
	yaml.Unmarshal([]byte(fmt.Sprintf(manifest, file.Name())), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithReleaseSettings(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
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
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithNoCapabilitySettingsEmptyDoc(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	manifest := `
it: should work
asserts:
  - hasDocuments:
      count: 0
    template: crd_backup.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithCapabilitySettings(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
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
    template: crd_backup.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithChartSettings(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
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
    template: deployment.yaml
  - equal:
      path: spec.template.spec.containers[0].image
      value: nginx:9999
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    documentIndex: 0
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithTestJobTemplateOk(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should work
template: deployment.yaml
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
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
  - deployment.yaml
  - configmap.yaml
asserts:
  - equal:
      path: kind
      value: Deployment   
    template: deployment.yaml
  - equal:
      path: kind
      value: ConfigMap   
    template: configmap.yaml
  - isNotNull:
      path: metadata.name
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
template: ingress.yaml
asserts:
  - failedTemplate:
      errorMessage: The externalPort is required
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, false, &results.TestJobResult{})
	// Write Buffer

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
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    documentIndex: 0
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})
	// Write Buffer

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
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
    template: deployment.yaml
`
	file, _ := ioutil.TempFile("", "testjob_test_TestRunJobWithValuesFile.yaml")
	file.WriteString("nameOverride: mary-jane")

	var tj TestJob
	yaml.Unmarshal([]byte(fmt.Sprintf(manifest, file.Name())), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

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
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
    template: crd_backup.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
    template: crd_backup.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
    template: crd_backup.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
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
    template: deployment.yaml
  - equal:
      path: spec.template.spec.containers[0].image
      value: nginx:9999
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithFailingTemplate(t *testing.T) {
	c, _ := loader.Load(testV3WithFailingTemplateChart)
	manifest := `
it: should work
template: configMap.yaml
asserts:
  - failedTemplate:
      errorMessage: no template "non-existing-named-template" associated with template "gotpl"
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

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
template: dummy.yaml
asserts:
  - failedTemplate:
      errorMessage: "values don't meet the specifications of the schema(s) in the following chart(s):\nwith-schema:\n- (root): image is required\n"
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, true, &results.TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.NotNil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}
