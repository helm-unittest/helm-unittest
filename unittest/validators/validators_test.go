package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	. "github.com/lrills/helm-unittest/unittest/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	yaml "gopkg.in/yaml.v2"
)

func TestEqualValidatorWhenOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := EqualValidator{"a.b[0].c", 123}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenNotOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := EqualValidator{"a.b[0].c", 321}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenFail(t *testing.T) {
	manifest := `
a:
  b:
    - c: 123
`
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := EqualValidator{"a.b[0]", map[interface{}]interface{}{"d": 321}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := EqualValidator{"a.b[0]", map[interface{}]interface{}{"c": 123}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := EqualValidator{"a.b.e", map[string]int{"d": 321}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := MatchRegexValidator{"a.b[0].c", "^hello"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenNotOk(t *testing.T) {
	manifest := `
a:
  b:
    - c: hello world
`
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := MatchRegexValidator{"a.b[0].c", "^foo"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenRegexCompileFail(t *testing.T) {
	data := common.K8sManifest{"a": "A"}

	v := MatchRegexValidator{"a", "+"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := MatchRegexValidator{"a", "^foo"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := MatchRegexValidator{"a.b[0].c", "^bar"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := MatchRegexValidator{"a.b[0].c", "^foo"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "foo bar"}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "hello bar"}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := ContainsValidator{"a.b", map[interface{}]interface{}{"e": "bar bar"}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "foo bar"}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := ContainsValidator{"a.b", common.K8sManifest{"d": "foo bar"}}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenNotOk(t *testing.T) {
	manifest := "a: 0"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenFail(t *testing.T) {
	manifest := "a: A"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected to be null, got:",
		"	A",
	}, diff)
}

func TestIsNullValidatorWhenNOTFail(t *testing.T) {
	manifest := "a:"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key := range data {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key := range data {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key, value := range data {
		validator := IsEmptyValidator{key}
		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)
		pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	for key, value := range data {
		validator := IsEmptyValidator{key}
		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)
		pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
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
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenNotOk(t *testing.T) {
	manifest := "kind: Service"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenFail(t *testing.T) {
	manifest := "kind: Pod"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsKindValidator{"Service"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected kind:	Service",
		"Actual:	Pod",
	}, diff)
}

func TestIsKindValidatorWhenNotFail(t *testing.T) {
	manifest := "kind: Pod"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to be kind:	Pod",
	}, diff)
}

func TestIsAPiVersionValidatorWhenOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsAPIVersionValidator{"v1"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPiVersionValidatorWhenNotOk(t *testing.T) {
	manifest := "apiVersion: v1"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsAPIVersionValidator{"v2"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPIVersionValidatorWhenFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsAPIVersionValidator{"v2"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected apiVersion:	v2",
		"Actual:	v1",
	}, diff)
}

func TestIsAPIVersionValidatorWhenNOTFail(t *testing.T) {
	manifest := "apiVersion: v1"
	data := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &data)

	v := IsAPIVersionValidator{"v1"}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to be apiVersion:	v1",
	}, diff)
}

func TestHasDocumentsValidatorOk(t *testing.T) {
	data := common.K8sManifest{}

	v := HasDocumentsValidator{2}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data, data}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorNotOk(t *testing.T) {
	data := common.K8sManifest{}

	v := HasDocumentsValidator{2}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data}, Negative: true})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorFail(t *testing.T) {
	data := common.K8sManifest{}

	v := HasDocumentsValidator{1}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data, data}})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected documents count:	1",
		"Actual:	2",
	}, diff)
}

func TestHasDocumentsValidatorNotFail(t *testing.T) {
	data := common.K8sManifest{}

	v := HasDocumentsValidator{2}
	pass, diff := v.Validate(&ValidateContext{Docs: []common.K8sManifest{data, data}, Negative: true})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected documents count NOT to be:	2",
	}, diff)
}

type mockSnapshotComparer struct {
	mock.Mock
}

func (m *mockSnapshotComparer) CompareToSnapshot(content interface{}) *snapshot.CompareResult {
	args := m.Called(content)
	return args.Get(0).(*snapshot.CompareResult)
}

func TestSnapshotValidatorWhenOk(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	v := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: true,
	})

	pass, diff := v.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenNotOk(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	v := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: false,
		Cached: "a:\n  b: c\n",
		New:    "x:\n  y: x\n",
	})

	pass, diff := v.Validate(&ValidateContext{
		Negative:         true,
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenFail(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	v := MatchSnapshotValidator{Path: "a"}

	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: false,
		Cached: "a:\n  b: c\n",
		New:    "x:\n  y: x\n",
	})

	pass, diff := v.Validate(&ValidateContext{
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected to match snapshot 0:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,3 +1,3 @@",
		"	-a:",
		"	-  b: c",
		"	+x:",
		"	+  y: x",
	}, diff)

	mockComparer.AssertExpectations(t)
}

func TestSnapshotValidatorWhenNotFail(t *testing.T) {
	data := common.K8sManifest{"a": "b"}
	v := MatchSnapshotValidator{Path: "a"}

	cached := "a:\n  b: c\n"
	mockComparer := new(mockSnapshotComparer)
	mockComparer.On("CompareToSnapshot", "b").Return(&snapshot.CompareResult{
		Passed: true,
		Cached: cached,
		New:    cached,
	})

	pass, diff := v.Validate(&ValidateContext{
		Negative:         true,
		Docs:             []common.K8sManifest{data},
		SnapshotComparer: mockComparer,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:	a",
		"Expected NOT to match snapshot 0:",
		"	a:",
		"	  b: c",
	}, diff)

	mockComparer.AssertExpectations(t)
}
