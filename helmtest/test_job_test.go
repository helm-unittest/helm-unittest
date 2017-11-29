package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalableFromYAML(t *testing.T) {
	a := assert.New(t)
	manifest := `
it: should do something
values:
  - values.yaml
set:
  a.b.c: ABC
  x.y.z: XYZ
asserts:
  - matchSnapshot:
    documentIndex: 1
    not: true
  - matchValue:
      path: a.b
      value: c
  - matchPattern:
      path: x.y
      pattern: /z/
`
	var tj TestJob
	err := yaml.Unmarshal([]byte(manifest), &tj)

	a.Nil(err)
	a.Equal(tj.Name, "should do something")
	a.Equal(tj.Values, []string{"values.yaml"})
	a.Equal(tj.Set, map[string]interface{}{
		"a.b.c": "ABC",
		"x.y.z": "XYZ",
	})
	a.Equal(tj.Assertions, []Assertion{
		Assertion{DocumentIndex: 1, Not: true},
		Assertion{DocumentIndex: 0, Not: false},
		Assertion{DocumentIndex: 0, Not: false},
	})
}
