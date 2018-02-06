package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestEqualValidatorWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualValidator{"a.b[0].c", 123}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualValidator{"a.b[0]", map[string]int{"d": 321}}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b[0]",
		"Expected:",
		"	d: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-d: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorWhenWrongPath(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualValidator{"a.b.e", map[string]int{"d": 321}}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	can't get [\"e\"] from a non map type:",
		"	- c: 123",
	}, diff)
}

func TestMatchRegexValidatorWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexValidator{"a.b[0].c", "^hello"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenRegexCompileFail(t *testing.T) {
	data := K8sManifest{"a": "A"}

	a := MatchRegexValidator{"a", "+"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	error parsing regexp: missing argument to repetition operator: `+`",
	}, diff)
}

func TestMatchRegexValidatorWhenNotString(t *testing.T) {
	manifest := `
a: 123.456
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexValidator{"a", "^foo"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	expect 'a' to be a string, got:",
		"	123.456",
	}, diff)
}

func TestMatchRegexValidatorWhenMatchFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: foo
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexValidator{"a.b[0].c", "^bar"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b[0].c",
		"Expected to Match:	^bar",
		"Actual:	foo",
	}, diff)
}

func TestContainsValidatorWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "foo bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsValidator{"a.b", K8sManifest{"e": "bar bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b",
		"Expected Contains:",
		"	- e: bar bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}

func TestMatchContainsValidatorWhenNotAnArray(t *testing.T) {
	manifest := `
a:
  b:
    c: hello world
    d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsValidator{"a.b", K8sManifest{"d": "foo bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	expect 'a.b' to be an array, got:",
		"	c: hello world",
		"	d: foo bar",
	}, diff)
}

func TestIsNullValidatorWhenOk(t *testing.T) {
	manifest := "a:"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenFail(t *testing.T) {
	manifest := "a: A"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected:	null",
		"Actual:",
		"	A",
	}, diff)
}

func TestIsEmptyValidatorWhenOk(t *testing.T) {
	manifest := `
a:
b: ""
c: 0
d: null
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsEmptyValidator{"a"}
	aPass, aDiff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, aPass)
	assert.Equal(t, []string{}, aDiff)

	b := IsEmptyValidator{"b"}
	bPass, bDiff := b.Validate([]K8sManifest{data}, 0)
	assert.True(t, bPass)
	assert.Equal(t, []string{}, bDiff)

	c := IsEmptyValidator{"c"}
	cPass, cDiff := c.Validate([]K8sManifest{data}, 0)
	assert.True(t, cPass)
	assert.Equal(t, []string{}, cDiff)

	d := IsEmptyValidator{"d"}
	dPass, dDiff := d.Validate([]K8sManifest{data}, 0)
	assert.True(t, dPass)
	assert.Equal(t, []string{}, dDiff)
}

func TestIsEmptyValidatorWhenFail(t *testing.T) {
	manifest := "a: 1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsEmptyValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected to be empty, got:",
		"	1",
	}, diff)
}

func TestIsKindValidatorWhenOk(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Pod"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenFail(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Service"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected kind:	Service",
		"Actual:	Pod",
	}, diff)
}

func TestIsAPiVersionValidatorWhenOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v1"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPIVersionValidatorWhenFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v2"}
	pass, diff := a.Validate([]K8sManifest{data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected apiVersion:	v2",
		"Actual:	v1",
	}, diff)
}

func TestHasDocumentsValidatorOk(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{2}
	pass, diff := a.Validate([]K8sManifest{data, data}, 0)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorFail(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{1}
	pass, diff := a.Validate([]K8sManifest{data, data}, 0)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected:	1",
		"Actual:	2",
	}, diff)
}
