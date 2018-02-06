package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"k8s.io/helm/pkg/chartutil"
)

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
	assertions := make([]Assertion, 2)
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

func TestRunJobOk(t *testing.T) {
	c, _ := chartutil.Load("../__fixtures__/basic")
	manifest := `
it: should work
asserts:
  - equal:
      path: kind
      value: Deployment
    file: deployment.yaml
  - matchRegex:
      path: metadata.name
      pattern: -basic$
    file: deployment.yaml
`
	var tj TestJob
	yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.Run(c)

	a := assert.New(t)
	a.Nil(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}

func TestRunJobWithAssertionFail(t *testing.T) {
	c, _ := chartutil.Load("../__fixtures__/basic")
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

	testResult := tj.Run(c)

	a := assert.New(t)
	a.Nil(testResult.ExecError)
	a.False(testResult.Passed)
	a.Equal(2, len(testResult.AssertsResult))
}
