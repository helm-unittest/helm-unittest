package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/stretchr/testify/assert"
)

// Tests for normalizing multiline objects with map[string]any and []any

var docWithArrayAny = `
items:
  - name: item1
    content: |

      Multi
      Line
      Content
  - name: item2
    data: simple
`

// TestEqualValidatorArrayAnyWithMultilineWhenOk verifies []any normalization
func TestEqualValidatorArrayAnyWithMultilineWhenOk(t *testing.T) {
	manifest := makeManifest(docWithArrayAny)
	validator := EqualValidator{Path: "items", Value: []any{
		map[string]any{
			"name":    "item1",
			"content": "Multi\nLine\nContent\n",
		},
		map[string]any{
			"name": "item2",
			"data": "simple",
		},
	}, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// TestEqualValidatorArrayAnyWithMultilineWhenFail verifies []any mismatch detection
func TestEqualValidatorArrayAnyWithMultilineWhenFail(t *testing.T) {
	manifest := makeManifest(docWithArrayAny)
	validator := EqualValidator{Path: "items", Value: []any{
		map[string]any{
			"name":    "item1",
			"content": "Wrong\nContent\n",
		},
		map[string]any{
			"name": "item2",
			"data": "simple",
		},
	}, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.NotEmpty(t, diff)
}

// Regression test for https://github.com/helm-unittest/helm-unittest/issues/826
// equal assertion fails on identical multi-line block scalar in ConfigMap data
// when the actual value has leading newlines (from YAML round-trip).
func TestEqualValidatorMapWithLeadingNewlineInValue(t *testing.T) {
	// Simulate a manifest where the data value has a leading newline
	// This happens after YAML round-trip in GetValueOfSetPath
	manifest := common.K8sManifest{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": common.K8sManifest{
			"name": "test-config",
		},
		"data": common.K8sManifest{
			"bus.lua": "\nTEST CONFIG\nWITH MULTILINE CONFIG\n",
		},
	}

	// The expected value without leading newline (as parsed from test YAML)
	expectedData := map[string]any{
		"bus.lua": "TEST CONFIG\nWITH MULTILINE CONFIG\n",
	}

	validator := EqualValidator{Path: "data", Value: expectedData, DecodeBase64: false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass, "equal should pass when only difference is leading newlines in map values: %v", diff)
}

// Regression test for https://github.com/helm-unittest/helm-unittest/issues/826
// Trailing spaces before newlines in map values should not cause failure.
func TestEqualValidatorMapWithTrailingSpacesInValue(t *testing.T) {
	manifest := common.K8sManifest{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": common.K8sManifest{
			"name": "test-config",
		},
		"data": common.K8sManifest{
			"bus.lua": "TEST CONFIG \nWITH MULTILINE CONFIG \n",
		},
	}

	expectedData := map[string]any{
		"bus.lua": "TEST CONFIG\nWITH MULTILINE CONFIG\n",
	}

	validator := EqualValidator{Path: "data", Value: expectedData, DecodeBase64: false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass, "equal should pass when only difference is trailing spaces in map values: %v", diff)
}

// Test that equal still fails when values are genuinely different
func TestEqualValidatorMapWithDifferentValuesStillFails(t *testing.T) {
	manifest := common.K8sManifest{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": common.K8sManifest{
			"name": "test-config",
		},
		"data": common.K8sManifest{
			"bus.lua": "DIFFERENT CONFIG\n",
		},
	}

	expectedData := map[string]any{
		"bus.lua": "TEST CONFIG\n",
	}

	validator := EqualValidator{Path: "data", Value: expectedData, DecodeBase64: false}
	pass, _ := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass, "equal should fail when values are genuinely different")
}

// Test with nested maps containing multi-line strings
func TestEqualValidatorNestedMapWithLeadingNewline(t *testing.T) {
	manifest := common.K8sManifest{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": common.K8sManifest{
			"name": "test-config",
		},
		"data": common.K8sManifest{
			"config.yaml": "\nkey: value\nnested:\n  data: test\n",
		},
	}

	expectedData := map[string]any{
		"config.yaml": "key: value\nnested:\n  data: test\n",
	}

	validator := EqualValidator{Path: "data", Value: expectedData, DecodeBase64: false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass, "equal should pass for nested map with leading newline: %v", diff)
}
