package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsNullValidatorWhenOk(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenArrayOk(t *testing.T) {
	doc := `
a:
  - b
`
	manifest := makeManifest(doc)

	v := IsNullValidator{"a[1]"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenNegativeAndOk(t *testing.T) {
	doc := "a: 0"
	manifest := makeManifest(doc)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullValidatorWhenFail(t *testing.T) {
	doc := "a: A"
	manifest := makeManifest(doc)

	log.SetLevel(log.DebugLevel)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a",
		"Expected to be null, got:",
		"	A",
	}, diff)
}

func TestIsNullValidatorWhenNegativeAndFail(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := IsNullValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a",
		"Expected NOT to be null, got:",
		"	null",
	}, diff)
}

func TestIsNullValidatorWhenInvalidIndex(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	validator := IsNullValidator{"a"}
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

func TestIsNullValidatorWhenInvalidPath(t *testing.T) {
	doc := "x:"
	manifest := makeManifest(doc)

	validator := IsNullValidator{"x.b"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"b\"] from a non map type:",
		"	null",
	}, diff)
}
