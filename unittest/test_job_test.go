package unittest_test

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/lrills/helm-unittest/unittest"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"

	v2util "k8s.io/helm/pkg/chartutil"
)

func makeTestJobResultSnapshotable(result *TestJobResult) *TestJobResult {
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
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithAssertionFail(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    file: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    file: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, &TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV2RunJobWithValueSet(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	manifest := `
it: should work
set:
  nameOverride: john-doe
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-john-doe
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithValuesFile(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	manifest := `
it: should work
values:
  - %s
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-mary-jane
    template: deployment.yaml
`
	file, _ := ioutil.TempFile("", "testjob_test_TestRunJobWithValuesFile.yaml")
	file.WriteString("nameOverride: mary-jane")

	var tj TestJob
	yaml.Unmarshal([]byte(fmt.Sprintf(manifest, file.Name())), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV2RunJobWithReleaseSetting(t *testing.T) {
	c, _ := v2util.Load("../__fixtures__/v2/basic")
	manifest := `
it: should work
release:
  name: my-release
asserts:
  - equal:
      path: metadata.name
      value: my-release-basic
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV2(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobOk(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    template: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithAssertionFail(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: WrongKind
    file: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: pattern-not-match
    file: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, &TestJobResult{})
	// Write Buffer

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestV3RunJobWithValueSet(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	manifest := `
it: should work
set:
  nameOverride: john-doe
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-john-doe
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithValuesFile(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	manifest := `
it: should work
values:
  - %s
asserts:
  - equal:
      path: metadata.name
      value: RELEASE-NAME-mary-jane
    template: deployment.yaml
`
	file, _ := ioutil.TempFile("", "testjob_test_TestRunJobWithValuesFile.yaml")
	file.WriteString("nameOverride: mary-jane")

	var tj TestJob
	yaml.Unmarshal([]byte(fmt.Sprintf(manifest, file.Name())), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}

func TestV3RunJobWithReleaseSetting(t *testing.T) {
	c, _ := loader.Load("../__fixtures__/v3/basic")
	manifest := `
it: should work
release:
  name: my-release
asserts:
  - equal:
      path: metadata.name
      value: my-release-basic
    template: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, &snapshot.Cache{}, &TestJobResult{})

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestJobResultSnapshotable(testResult))

	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(1, len(testResult.AssertsResult))
}
