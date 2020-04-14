package validators_test

import (
	"testing"

	. "github.com/lrills/helm-unittest/unittest/validators"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/stretchr/testify/assert"
)

var docToTestContains = `
a:
  b:
    - c: hello world
    - d: foo bar
`

func TestContainsValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{"a.b", map[interface{}]interface{}{"d": "hello bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: bar bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	var docToTestContains = `
a:
  b:
    - c: hello world
`
	manifest2 := makeManifest(docToTestContains)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  manifests,
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifests := []common.K8sManifest{manifest1, manifest1}

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"e": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  manifests,
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}

func TestContainsValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		"a.b",
		map[interface{}]interface{}{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected NOT to contain:",
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}

func TestMatchContainsValidatorWhenNotAnArray(t *testing.T) {
	manifestDocNotArray := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotArray)

	validator := ContainsValidator{"a.b", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	expect 'a.b' to be an array, got:",
		"	c: hello world",
		"	d: foo bar",
	}, diff)
}

func TestContainsValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{"a.b", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest},
		Index: 2,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	documentIndex 2 out of range",
	}, diff)
}

func TestContainsValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{"a.b.e", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"e\"] from a non map type:",
		"	- c: hello world",
		"	- d: foo bar",
	}, diff)
}
