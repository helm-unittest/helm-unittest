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

var docToTestEqualMultiplePaths = `
a:
  b: 1
  c: 1
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
	assert.Equal(t, []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	unable to decode base64 expected content Line1 ", "	Line2"}, diff)
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
		map[any]any{"d": 321},
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
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
		map[string]any{"c": 321},
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"ValuesIndex:	0",
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
		map[string]any{"c": 321},
		false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
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
		"ValuesIndex:	0",
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

	v := EqualValidator{"a.b[0]", map[string]any{"c": 123}, false}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
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

func TestEqualValidatorWhenUnknownPath(t *testing.T) {
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

func TestEqualValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{"a.b[5]", map[string]int{"d": 321}, false}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenUnknownPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{"a.b[5]", map[string]int{"d": 321}, false}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}

func TestEqualValidatorWhenOkWithMultiplePaths(t *testing.T) {
	manifest := makeManifest(docToTestEqualMultiplePaths)
	validator := EqualValidator{"a.*", 1, false}

	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWithMultiplePathsFailFast(t *testing.T) {
	manifest := makeManifest(docToTestEqualMultiplePaths)
	validator := EqualValidator{"a.*", 2, true}

	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"ValuesIndex:\t0",
		"Path:\ta.*",
		"Expected to equal:",
		"\t2",
		"Actual:",
		"\t1",
		"Diff:",
		"\t--- Expected",
		"\t+++ Actual",
		"\t@@ -1,2 +1,2 @@",
		"\t-2",
		"\t+1"}, diff)
}

func TestEqualValidatorWhenNoManifestFail(t *testing.T) {
	validator := EqualValidator{"a.b[0].c", 123, false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\ta.b[0].c",
		"Expected to equal:",
		"\t123",
		"Actual:",
		"\tno manifest found",
		"Diff:",
		"\t--- Expected",
		"\t+++ Actual",
		"\t@@ -1,2 +1,2 @@",
		"\t-123",
		"\t+no manifest found"}, diff)
}

func TestEqualValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := EqualValidator{"a.b[0].c", 123, false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
