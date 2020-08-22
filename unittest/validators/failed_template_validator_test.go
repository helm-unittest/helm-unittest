package validators_test

import (
	"testing"

	. "github.com/lrills/helm-unittest/unittest/validators"

	"github.com/lrills/helm-unittest/unittest/common"
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

func TestMultiManfiestFailedTemplateValidatorWhenOk(t *testing.T) {
	manifest1 := makeManifest(failedTemplate)
	manifest2 := makeManifest(failedTemplate)
	validator := FailedTemplateValidator{"A field should be required"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest1, manifest2},
		Index: -1,
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

	validator := FailedTemplateValidator{"A field should not be required"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected:",
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
