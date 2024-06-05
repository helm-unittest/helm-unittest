package validators_test

import (
	"errors"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var failedTemplate = `
raw: A field should be required
`

func TestFailedTemplateValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{"A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{manifest},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{"A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{manifest},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenEmptyFail(t *testing.T) {
	tests := []struct {
		name      string
		validator FailedTemplateValidator
		expected  []string
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{"A field should not be required"},
			expected:  []string{"Expected to equal:", "\tA field should not be required", "Actual:", "\tNo failed document"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
			expected:  []string{"Expected to equal:", "\t", "Actual:", "\tNo failed document"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{},
				Negative: false,
			})

			assert.False(t, pass)
			assert.Equal(t, tt.expected, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenEmptyNegativeAndOk(t *testing.T) {
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "test case 1: with error message",
			validator: FailedTemplateValidator{"A field should not be required"},
		},
		{
			name:      "test case 2: empty error message",
			validator: FailedTemplateValidator{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:     []common.K8sManifest{},
				Negative: true,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(failedTemplate)

	log.SetLevel(log.DebugLevel)

	tests := []struct {
		name      string
		validator FailedTemplateValidator
		expected  []string
	}{
		{
			name:      "test case 1: incorrect error message",
			validator: FailedTemplateValidator{"A field should not be required"},
			expected: []string{
				"DocumentIndex:	0",
				"Expected to equal:",
				"	A field should not be required",
				"Actual:",
				"	A field should be required",
			},
		},
		{
			name:      "test case 2: empty error message",
			validator: FailedTemplateValidator{},
			expected:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})

			if len(tt.expected) > 0 {
				assert.False(t, pass)
			} else {
				assert.True(t, pass)
			}

			assert.Equal(t, tt.expected, diff)
		})
	}
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

func TestFailedTemplateValidatorWhenRenderError(t *testing.T) {
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "Test case 1: with error message",
			validator: FailedTemplateValidator{"values don't meet the specifications of the schema(s)"},
		},
		{
			name:      "Test case 2: empty error message",
			validator: FailedTemplateValidator{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pass, diff := tt.validator.Validate(&ValidateContext{
				Docs:        []common.K8sManifest{},
				RenderError: errors.New(tt.validator.ErrorMessage),
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}
