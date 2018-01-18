package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestEqualAsserterWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualAsserter{"a.b[0].c", 123}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, diff, "")
}

func TestEqualAsserterWhenFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualAsserter{"a.b[0]", map[string]int{"d": 321}}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Path: a.b[0]
	Expected:
		d: 321
	Actual:
		c: 123
	Diff:
		--- Expected
		+++ Actual
		@@ -1,2 +1,2 @@
		-d: 321
		+c: 123
`, diff)
}

func TestEqualAsserterWhenWrongPath(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualAsserter{"a.b.e", map[string]int{"d": 321}}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Error:
		can't get ["e"] from a non map type:
		- c: 123
`, diff)
}

func TestMatchRegexAsserterWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexAsserter{"a.b[0].c", "^hello"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, diff, "")
}

func TestMatchRegexAsserterWhenRegexCompileFail(t *testing.T) {
	data := K8sManifest{"a": "A"}

	a := MatchRegexAsserter{"a", "+"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, diff, `
	Error:
		error parsing regexp: missing argument to repetition operator: `+"`+`\n")
}

func TestMatchRegexAsserterWhenNotString(t *testing.T) {
	manifest := `
a: 123.456
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexAsserter{"a", "^foo"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, diff, `
	Error:
		expect 'a' to be a string, got:
		123.456
`)
}

func TestMatchRegexAsserterWhenMatchFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: foo
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexAsserter{"a.b[0].c", "^bar"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Path: a.b[0].c
	Expected to Match: ^bar
	Actual: foo
`, diff)
}

func TestContainsAsserterWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsAsserter{"a.b", map[interface{}]interface{}{"d": "foo bar"}}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, "", diff)
}

func TestContainsAsserterWhenFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsAsserter{"a.b", K8sManifest{"e": "bar bar"}}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Path: a.b
	Expected Contains:
		- e: bar bar
	Actual:
		- c: hello world
		- d: foo bar
`, diff)
}

func TestMatchContainsAsserterWhenNotAnArray(t *testing.T) {
	manifest := `
a:
  b:
    c: hello world
    d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsAsserter{"a.b", K8sManifest{"d": "foo bar"}}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Error:
		expect 'a.b' to be an array, got:
		c: hello world
		d: foo bar
`, diff)
}

func TestIsNullAsserterWhenOk(t *testing.T) {
	manifest := "a:"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullAsserter{"a"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, "", diff)
}

func TestIsNullAsserterWhenFail(t *testing.T) {
	manifest := "a: A"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullAsserter{"a"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Path: a
	Expected: null
	Actual:
		A
`, diff)
}

func TestIsEmptyAsserterWhenOk(t *testing.T) {
	manifest := `
a:
b: ""
c: 0
d: null
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsEmptyAsserter{"a"}
	aPass, aDiff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, aPass)
	assert.Equal(t, "", aDiff)

	b := IsEmptyAsserter{"b"}
	bPass, bDiff := b.Assert([]K8sManifest{data}, 0)
	assert.True(t, bPass)
	assert.Equal(t, "", bDiff)

	c := IsEmptyAsserter{"c"}
	cPass, cDiff := c.Assert([]K8sManifest{data}, 0)
	assert.True(t, cPass)
	assert.Equal(t, "", cDiff)

	d := IsEmptyAsserter{"d"}
	dPass, dDiff := d.Assert([]K8sManifest{data}, 0)
	assert.True(t, dPass)
	assert.Equal(t, "", dDiff)
}

func TestIsEmptyAsserterWhenFail(t *testing.T) {
	manifest := "a: 1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsEmptyAsserter{"a"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Path: a
	Expected to be empty, got:
		1
`, diff)
}

func TestIsKindAsserterWhenOk(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindAsserter{"Pod"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, "", diff)
}

func TestIsKindAsserterWhenFail(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindAsserter{"Service"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Expected 'kind': Service
	Actual: Pod
`, diff)
}

func TestIsAPiVersionAsserterWhenOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionAsserter{"v1"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, "", diff)
}

func TestIsAPIVersionAsserterWhenFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionAsserter{"v2"}
	pass, diff := a.Assert([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Expected 'apiVersion': v2
	Actual: v1
`, diff)
}

func TestHasDocumentsAsserterOk(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsAsserter{2}
	pass, diff := a.Assert([]K8sManifest{data, data}, 0)
	assert.True(t, pass)
	assert.Equal(t, "", diff)
}

func TestHasDocumentsAsserterFail(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsAsserter{1}
	pass, diff := a.Assert([]K8sManifest{data, data}, 0)
	assert.False(t, pass)
	assert.Equal(t, `
	Expected: 1
	Actual: 2
`, diff)
}
