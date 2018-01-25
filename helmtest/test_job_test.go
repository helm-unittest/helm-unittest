package helmtest_test

import (
	"bytes"
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"k8s.io/helm/pkg/chartutil"
)

func TestUnmarshalableFromYAML(t *testing.T) {
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

	a := assert.New(t)
	var buf bytes.Buffer
	pass, err := tj.Run(c, &buf)

	a.Nil(err)
	a.True(pass)
	a.Equal("", buf.String())
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

	a := assert.New(t)
	var buf bytes.Buffer
	pass, err := tj.Run(c, &buf)

	a.Nil(err)
	a.False(pass)
	a.Equal(`
"should work": failed
- asserts[0] `+"`equal`"+` fail:

	Path: kind
	Expected:
		WrongKind
	Actual:
		Deployment
	Diff:
		--- Expected
		+++ Actual
		@@ -1,2 +1,2 @@
		-WrongKind
		+Deployment

- asserts[1] `+"`matchRegex`"+` fail:

	Path: metadata.name
	Expected to Match: pattern-not-match
	Actual: RELEASE_NAME-basic
`, buf.String())
}
