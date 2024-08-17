package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestExistsValidatorWhenOk(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := ExistsValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenArrayOk(t *testing.T) {
	doc := `
a:
  - b
`
	manifest := makeManifest(doc)

	v := ExistsValidator{"a[0]"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenNegativeAndOk(t *testing.T) {
	doc := "a: 0"
	manifest := makeManifest(doc)

	v := ExistsValidator{"b"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenFail(t *testing.T) {
	doc := "a: A"
	manifest := makeManifest(doc)

	log.SetLevel(log.DebugLevel)

	v := ExistsValidator{"b"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	b expected to exists",
	}, diff)
}

func TestExistsValidatorWhenNegativeAndFail(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := ExistsValidator{"a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a expected to NOT exists",
	}, diff)
}

func TestExistsValidatorWhenInvalidPath(t *testing.T) {
	doc := "x:"
	manifest := makeManifest(doc)

	validator := ExistsValidator{"x[b]"}
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
