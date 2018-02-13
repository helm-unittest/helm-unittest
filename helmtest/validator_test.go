package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenNotOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualValidator{"a.b[0].c", 321}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
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

	a := EqualValidator{"a.b[0]", map[interface{}]interface{}{"d": 321}}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
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

func TestEqualValidatorWhenNotFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := EqualValidator{"a.b[0]", map[interface{}]interface{}{"c": 123}}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b[0]",
		"Expected NOT to equal:",
		"	c: 123",
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenNotOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexValidator{"a.b[0].c", "^foo"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenRegexCompileFail(t *testing.T) {
	data := K8sManifest{"a": "A"}

	a := MatchRegexValidator{"a", "+"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b[0].c",
		"Expected to match:	^bar",
		"Actual:	foo",
	}, diff)
}

func TestMatchRegexValidatorWhenMatchNotFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: foo
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := MatchRegexValidator{"a.b[0].c", "^foo"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b[0].c",
		"Expected NOT to match:	^foo",
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenNotOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "hello bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
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

	a := ContainsValidator{"a.b", map[interface{}]interface{}{"e": "bar bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b",
		"Expected to contain:",
		"	- e: bar bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}

func TestContainsValidatorWhenNotFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
    - d: foo bar
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "foo bar"}}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a.b",
		"Expected NOT to contain:",
		"	- d: foo bar",
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
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
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenNotOk(t *testing.T) {
	manifest := "a: 0"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenFail(t *testing.T) {
	manifest := "a: A"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected to be null, got:",
		"	A",
	}, diff)
}

func TestIsNullValidatorWhenNOTFail(t *testing.T) {
	manifest := "a:"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsNullValidator{"a"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected NOT to be null, got:",
		"	null",
	}, diff)
}

func TestIsEmptyValidatorWhenOk(t *testing.T) {
	manifest := `
a:
b: ""
c: 0
d: null
e: []
f: {}
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key := range data {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate([]K8sManifest{data}, 0, false)
		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsEmptyValidatorWhenNotOk(t *testing.T) {
	manifest := `
a: {a: A}
b: "b"
c: 1
d: [d]
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key := range data {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate([]K8sManifest{data}, 0, true)
		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsEmptyValidatorWhenFail(t *testing.T) {
	manifest := `
a: {a: A}
b: "b"
c: 1
d: [d]
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key, value := range data {
		validator := IsEmptyValidator{key}
		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)
		pass, diff := validator.Validate([]K8sManifest{data}, 0, false)
		assert.False(t, pass)
		assert.Equal(t, []string{
			"Path:	" + key,
			"Expected to be empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsEmptyValidatorWhenNotFail(t *testing.T) {
	manifest := `
a:
b: ""
c: 0
d: null
e: []
f: {}
`
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key, value := range data {
		validator := IsEmptyValidator{key}
		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)
		pass, diff := validator.Validate([]K8sManifest{data}, 0, true)
		assert.False(t, pass)
		assert.Equal(t, []string{
			"Path:	" + key,
			"Expected NOT to be empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsKindValidatorWhenOk(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Pod"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenNotOk(t *testing.T) {
	manifest := "kind: Service"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Pod"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenFail(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Service"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected kind:	Service",
		"Actual:	Pod",
	}, diff)
}

func TestIsKindValidatorWhenNotFail(t *testing.T) {
	manifest := "kind: Pod"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsKindValidator{"Pod"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to be kind:	Pod",
	}, diff)
}

func TestIsAPiVersionValidatorWhenOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v1"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPiVersionValidatorWhenNotOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v2"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPIVersionValidatorWhenFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v2"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected apiVersion:	v2",
		"Actual:	v1",
	}, diff)
}

func TestIsAPIVersionValidatorWhenNOTFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	a := IsAPIVersionValidator{"v1"}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to be apiVersion:	v1",
	}, diff)
}

func TestHasDocumentsValidatorOk(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{2}
	pass, diff := a.Validate([]K8sManifest{data, data}, 0, false)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorNotOk(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{2}
	pass, diff := a.Validate([]K8sManifest{data}, 0, true)
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorFail(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{1}
	pass, diff := a.Validate([]K8sManifest{data, data}, 0, false)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected documents count:	1",
		"Actual:	2",
	}, diff)
}

func TestHasDocumentsValidatorNotFail(t *testing.T) {
	data := K8sManifest{}

	a := HasDocumentsValidator{2}
	pass, diff := a.Validate([]K8sManifest{data, data}, 0, true)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected documents count NOT to be:	2",
	}, diff)
}
