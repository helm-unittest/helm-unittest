package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
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

func TestIsKindValidatorWhenNegativeAndFailFast(t *testing.T) {
	doc := "kind: Pod"
	manifest := makeManifest(doc)

	v := IsKindValidator{"Pod"}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true},
	)
	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to be kind:",
		"	Pod",
	}, diff)
}

func TestIsKindValidatorWhenNoManifestFail(t *testing.T) {
	v := IsKindValidator{"Service"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected to be kind:",
		"\tService",
		"Actual:",
		"\tno manifest found",
	}, diff)
}

func TestIsKindValidatorWhenNoManifestNegativeOk(t *testing.T) {
	v := IsKindValidator{"Service"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
