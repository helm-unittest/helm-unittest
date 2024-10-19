package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestType = `
a:
  b:
    - c: 123
  e: |
    Line1
    Line2
`

func TestTypeValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestType)
	validator := IsTypeValidator{"a.b[0].c", "int"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorMultiLineWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestType)
	validator := IsTypeValidator{"a.e", "string"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestType)

	validator := IsTypeValidator{"a.b[0].c", "string"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	log.SetLevel(log.DebugLevel)

	validator := IsTypeValidator{"a.b[0]", "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0]",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	map[string]interface {}",
	}, diff)
}

func TestTypeValidatorWhenFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	log.SetLevel(log.DebugLevel)

	validator := IsTypeValidator{"a.b[0]", "int"}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0]",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	map[string]interface {}",
	}, diff)
}

func TestTypeValidatorMultiManifestWhenFail(t *testing.T) {
	correctDoc := `
a:
  b:
    - c: "123"
`
	manifest1 := makeManifest(correctDoc)
	manifest2 := makeManifest(docToTestType)

	validator := IsTypeValidator{"a.b[0].c", "string"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"Path:	a.b[0].c",
		"Expected to be of type:",
		"	string",
		"Actual:",
		"	int",
	}, diff)
}

func TestTypeValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	validator := IsTypeValidator{"a.e", "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.e",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	string",
		"DocumentIndex:	1",
		"Path:	a.e",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	string",
	}, diff)
}

func TestTypeValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{"a.e", "string"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.e",
		"Expected NOT to be of type:",
		"	string",
		"Actual:",
		"	string",
	}, diff)
}

func TestTypeValidatorWhenWrongPath(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{"a.b[e]", "int"}
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

func TestTypeValidatorWhenWrongPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{"a.b[e]", "int"}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestTypeValidatorWhenUnkownPath(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{"a.b[5]", "string"}
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

func TestTypeValidatorWhenUnkownPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{"a.b[5]", "string"}
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
