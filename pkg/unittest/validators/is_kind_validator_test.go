package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsKindValidatorWhenOk(t *testing.T) {
	doc := "kind: Pod"
	manifest := makeManifest(doc)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenNegativeAndOk(t *testing.T) {
	doc := "kind: Service"
	manifest := makeManifest(doc)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsKindValidatorWhenFail(t *testing.T) {
	doc := "kind: Pod"
	manifest := makeManifest(doc)

	log.SetLevel(log.DebugLevel)

	v := IsKindValidator{"Service"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected to be kind:",
		"	Service",
		"Actual:",
		"	Pod",
	}, diff)
}

func TestIsKindValidatorWhenNegativeAndFail(t *testing.T) {
	doc := "kind: Pod"
	manifest := makeManifest(doc)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true},
	)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to be kind:",
		"	Pod",
	}, diff)
}

func TestIsKindValidatorWhenInvalidIndex(t *testing.T) {
	doc := "kind: Pod"
	manifest := makeManifest(doc)

	validator := IsKindValidator{"Pod"}
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
