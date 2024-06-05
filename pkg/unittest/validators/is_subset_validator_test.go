package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestIsSubset = `
a:
  b:
    c: hello world
    d: foo bar
    x: baz
`

func TestIsSubsetValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"d": "foo bar", "x": "baz"}}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"d": "hello bar", "c": "hello world"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	log.SetLevel(log.DebugLevel)

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: bar bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestIsSubset)
	extraDoc := `
a:
  b:
    c: hello world
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: manifests,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	d: foo bar",
		"Actual:",
		"	c: hello world",
	}, diff)
}

func TestIsSubsetValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest1 := makeManifest(docToTestIsSubset)
	manifests := []common.K8sManifest{manifest1, manifest1}

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"e": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: manifests,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
		"DocumentIndex:	1",
		"Path:	a.b",
		"Expected to contain:",
		"	e: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		"a.b",
		map[string]interface{}{"d": "foo bar"},
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
		"	d: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorWhenNotAnObject(t *testing.T) {
	manifestDocNotObject := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotObject)

	validator := IsSubsetValidator{"a.b.c", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	expect 'a.b.c' to be an object, got:",
		"	hello world",
	}, diff)
}

func TestIsSubsetValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest("a::error")

	validator := IsSubsetValidator{"a[b]", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestIsSubsetValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest("a::error")

	validator := IsSubsetValidator{"a[5]", common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a[5]",
	}, diff)
}
