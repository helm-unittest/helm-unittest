package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
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

var docToTestEqualWithBase64 = `
a: MTIz
b: TGluZTEgCkxpbmUyCg==
`

func TestEqualValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.b[0].c", 123, false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorMultiLineWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.e", "Line1\nLine2\n", false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWithBase64WhenNOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{"a.e", "Line1\nLine2\n", true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:	0", "Error:", "	unable to decode base64 expected content Line1 ", "	Line2"}, diff)
}

func TestEqualValidatorWithBase64WhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqualWithBase64)
	validator := EqualValidator{"a", "123", true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorMultiLineWithBase64WhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqualWithBase64)
	validator := EqualValidator{"b", "Line1\nLine2\n", true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	validator := EqualValidator{"a.b[0].c", 321, false}
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
		false,
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
		map[string]interface{}{"c": 321},
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
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
		map[string]interface{}{"c": 321},
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
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

	v := EqualValidator{"a.b[0]", map[string]interface{}{"c": 123}, false}
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

	v := EqualValidator{"a.b[e]", map[string]int{"d": 321}, false}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestEqualValidatorWhenUnkownPath(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{"a.b[5]", map[string]int{"d": 321}, false}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}
