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
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
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
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
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
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
			expected:  []string{"Expected to equal:", "\tA field should not be required", "Actual:", "\tNo failed document"},
		},
		{
			name:      "test case 2: with empty error message",
			validator: FailedTemplateValidator{},
			expected:  []string{"Expected to equal:", "\t", "Actual:", "\tNo failed document"},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
			expected:  []string{"Expected to match:", "\tshould not be", "Actual:", "\tNo failed document"},
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
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
		},
		{
			name:      "test case 2: empty error message",
			validator: FailedTemplateValidator{},
		},
		{
			name:      "test case 3: with error message that match substring",
			validator: FailedTemplateValidator{ErrorPattern: "should not be"},
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
			validator: FailedTemplateValidator{ErrorMessage: "A field should not be required"},
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
		{
			name:      "test case 3: incorrect error message",
			validator: FailedTemplateValidator{ErrorPattern: "should not be required"},
			expected: []string{
				"DocumentIndex:	0",
				"Expected to match:",
				"	should not be required",
				"Actual:",
				"	A field should be required",
			},
		},
		{
			name:      "test case 4: with error message that give compile pattern error",
			validator: FailedTemplateValidator{ErrorPattern: "+"},
			expected:  []string{"DocumentIndex:\t0", "Expected to match:", "\t+", "Actual:", "\tA field should be required"},
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
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}
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
	testRenderError := errors.New("values don't meet the specifications of the schema(s)")
	tests := []struct {
		name      string
		validator FailedTemplateValidator
	}{
		{
			name:      "Test case 1: with error pattern",
			validator: FailedTemplateValidator{ErrorPattern: "schema"},
		},
		{
			name:      "Test case 2: with error message",
			validator: FailedTemplateValidator{ErrorMessage: "values don't meet the specifications of the schema(s)"},
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
				RenderError: testRenderError,
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestFailedTemplateValidatorWhenErrorMessageAndErrorPatternSet(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required", ErrorPattern: "pattern is set"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	single attribute 'errorMessage' or 'errorPattern' supported at the same time",
	}, diff)
}

func TestFailedTemplateValidatorShowsAllErrors(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}

	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Expected NOT to equal:",
		"\tA field should be required",
		"DocumentIndex:\t1",
		"Expected NOT to equal:",
		"\tA field should be required",
	}, diff)
}

func TestFailedTemplateValidatorFailFast(t *testing.T) {
	manifest := makeManifest(failedTemplate)
	v := FailedTemplateValidator{ErrorMessage: "A field should be required"}

	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Expected NOT to equal:",
		"	A field should be required",
	}, diff)
}

func TestFailedTemplateValidator_ErrorPattern_SpecialCharactersAndEscapes_OK(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)

	cases := []struct {
		name    string
		pattern string
	}{
		{
			name:    "pattern with backticks and escapes",
			pattern: "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` \\(root\\)",
		},
		{
			name:    "pattern with backticks",
			pattern: "`runAsNonRoot`",
		},
		{
			name:    "pattern with escape",
			pattern: "\\(root\\)",
		},
		{
			name:    "pattern contains metacharacters without escape",
			pattern: "(root)",
		},
		{
			name:    "pattern without escape",
			pattern: "(root)",
		},
		{
			name:    "pattern with meta characters and no explicit escape handling",
			pattern: "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			v := FailedTemplateValidator{ErrorPattern: tt.pattern}
			pass, _ := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})
			assert.True(t, pass)
		})
	}
}

func TestFailedTemplateValidator_ErrorPattern_SpecialCharactersAndEscapes_Diff(t *testing.T) {
	var template = "raw: |-\n    " + "`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)"
	manifest := makeManifest(template)

	cases := []struct {
		name    string
		pattern string
		diff    interface{}
	}{
		{
			name:    "pattern with incorrect regex escape",
			pattern: `\(root)`,
			diff: []string{
				"DocumentIndex:\t0",
				"Expected to match:",
				"\t\\(root)", "Actual:",
				"\t`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
			},
		},
		{
			name:    "pattern with incorrect regex escape",
			pattern: `\\`,
			diff: []string{
				"DocumentIndex:\t0",
				"Expected to match:",
				"\t\\\\", "Actual:",
				"\t`runAsNonRoot` is set to `true` but `runAsUser` is set to `0` (root)",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			v := FailedTemplateValidator{ErrorPattern: tt.pattern}
			_, diff := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})
			assert.Equal(t, diff, tt.diff)
		})
	}
}
