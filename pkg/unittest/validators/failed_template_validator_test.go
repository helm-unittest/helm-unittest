package validators_test

import (
	"errors"
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var failedTemplate = `
raw: A field should be required
`

func TestFailedTemplateValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	validator := FailedTemplateValidator{"A field should be required"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestFailedTemplateValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)

	validator := FailedTemplateValidator{"A field should not be required"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestFailedTemplateValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(failedTemplate)

	log.SetLevel(log.DebugLevel)

	validator := FailedTemplateValidator{"A field should not be required"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected to equal:",
		"	A field should not be required",
		"Actual:",
		"	A field should be required",
	}, diff)
}

func TestFailedTemplateValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(failedTemplate)

	v := FailedTemplateValidator{"A field should be required"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to equal:",
		"	A field should be required",
	}, diff)
}

func TestFailedTemplateValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	validator := FailedTemplateValidator{"A field should be required"}
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

func TestFailedTemplateValidatorWhenRenderError(t *testing.T) {
	validator := FailedTemplateValidator{"values don't meet the specifications of the schema(s)"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:        []common.K8sManifest{},
		Index:       -1,
		RenderError: errors.New("values don't meet the specifications of the schema(s)"),
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
