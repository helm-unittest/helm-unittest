package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsAPiVersionValidatorWhenOk(t *testing.T) {
	doc := "apiVersion: v1"
	manifest := makeManifest(doc)

	validator := IsAPIVersionValidator{"v1"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPiVersionValidatorWhenNegativeAndOk(t *testing.T) {
	doc := "apiVersion: v1"
	manifest := makeManifest(doc)

	validator := IsAPIVersionValidator{"v2"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsAPIVersionValidatorWhenFail(t *testing.T) {
	doc := "apiVersion: v1"
	manifest := makeManifest(doc)

	log.SetLevel(log.DebugLevel)

	validator := IsAPIVersionValidator{"v2"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected to be apiVersion:",
		"	v2",
		"Actual:",
		"	v1",
	}, diff)
}

func TestIsAPIVersionValidatorWhenNegativeAndFail(t *testing.T) {
	doc := "apiVersion: v1"
	manifest := makeManifest(doc)

	validator := IsAPIVersionValidator{"v1"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to be apiVersion:",
		"	v1",
	}, diff)
}
