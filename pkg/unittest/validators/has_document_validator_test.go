package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestHasDocumentsValidatorOk(t *testing.T) {
	data := common.K8sManifest{}

	validator := HasDocumentsValidator{2}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{data, data},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorWhenNegativeAndOk(t *testing.T) {
	data := common.K8sManifest{}

	validator := HasDocumentsValidator{2}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{data},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestHasDocumentsValidatorWhenFail(t *testing.T) {
	data := common.K8sManifest{}

	log.SetLevel(log.DebugLevel)

	validator := HasDocumentsValidator{1}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{data, data},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected documents count to be:",
		"	1",
		"Actual:",
		"	2",
	}, diff)
}

func TestHasDocumentsValidatorWhenNegativeAndFail(t *testing.T) {
	data := common.K8sManifest{}

	validator := HasDocumentsValidator{2}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{data, data},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT documents count to be:",
		"	2",
	}, diff)
}
