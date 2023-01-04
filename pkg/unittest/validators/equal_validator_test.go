package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestEqual = `
a:
  b:
    - c: 123
  e: |
    Line1 
    Line2
`

func TestEqualValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.b[0].c", 123}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorMultiLineWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.e", "Line1\nLine2\n"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	validator := EqualValidator{"a.b[0].c", 321}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	log.SetLevel(log.DebugLevel)

	validator := EqualValidator{
		"a.b[0]",
		map[interface{}]interface{}{"d": 321},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
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

func TestEqualValidatorMultiManifestWhenFail(t *testing.T) {
	correctDoc := `
a:
  b:
    - c: 321
`
	manifest1 := makeManifest(correctDoc)
	manifest2 := makeManifest(docToTestEqual)

	validator := EqualValidator{
		"a.b[0]",
		map[interface{}]interface{}{"c": 321},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest1, manifest2},
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	validator := EqualValidator{
		"a.b[0]",
		map[interface{}]interface{}{"c": 321},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest, manifest},
		Index: -1,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
		"DocumentIndex:	1",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{"a.b[0]", map[interface{}]interface{}{"c": 123}}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0]",
		"Expected NOT to equal:",
		"	c: 123",
	}, diff)
}

func TestEqualValidatorWhenWrongPath(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{"a.b.e", map[string]int{"d": 321}}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"e\"] from a non map type:",
		"	- c: 123",
	}, diff)
}

func TestEqualValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.b[0].c", 123}
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
