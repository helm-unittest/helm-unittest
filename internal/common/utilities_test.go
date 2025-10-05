package common_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestNewYamlNode(t *testing.T) {
	node := NewYamlNode()
	assert.NotNil(t, node)
	assert.NotNil(t, node.Node)
}

func TestYamlNewDecoder(t *testing.T) {
	reader := strings.NewReader("test: value")
	decoder := YamlNewDecoder(reader)
	assert.NotNil(t, decoder)
}

func TestYamlNewEncoder(t *testing.T) {
	var buf bytes.Buffer
	encoder := YamlNewEncoder(&buf)
	assert.NotNil(t, encoder)
}

func TestYamlToJson(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "simple key-value",
			input:    "name: test\nversion: 1.0",
			expected: `{"name":"test","version":1}`,
		},
		{
			name:     "nested object",
			input:    "metadata:\n  name: test\n  namespace: default",
			expected: `{"metadata":{"name":"test","namespace":"default"}}`,
		},
		{
			name:     "array",
			input:    "items:\n- name: item1\n- name: item2",
			expected: `{"items":[{"name":"item1"},{"name":"item2"}]}`,
		},
		{
			name:     "empty yaml",
			input:    "",
			expected: `null`,
		},
		{
			name:     "simple string",
			input:    "hello world",
			expected: `"hello world"`,
		},
		{
			name:        "invalid yaml",
			input:       "invalid:\n  - yaml\n- structure",
			expectError: true,
		},
		{
			name:     "boolean and number",
			input:    "enabled: true\ncount: 42",
			expected: `{"count":42,"enabled":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := YamlToJson(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(result))
			}
		})
	}
}

func TestTrustedUnmarshalYml(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]any
		expectError bool
	}{
		{
			name:  "valid yaml to map",
			input: "name: test\nversion: 1.0",
			expected: map[string]any{
				"name":    "test",
				"version": 1.0,
			},
		},
		{
			name: "valid yaml to map",
			input: `
---
# Source: full-snapshot/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: release-name-overrides
data:
  overrides.json: |

    {}
`,
			expected: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": "release-name-overrides",
				},
				"data": map[string]any{
					"overrides.json": "\n{}\n",
				},
			},
		},
		{
			name:        "empty yaml",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid yaml",
			input:       "invalid:\n  - yaml\n- structure",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Defer a function to recover from panic
			defer func() {
				if r := recover(); r != nil {
					// We successfully recovered from panic
					assert.True(t, tt.expectError)
				}
			}()
			result := TrustedUnmarshalYAML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrustedMarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    string
		expectError bool
	}{
		{
			name:     "simple map",
			input:    map[string]any{"name": "test", "version": "1.0"},
			expected: "name: test\nversion: \"1.0\"\n",
		},
		{
			name:     "nested map",
			input:    map[string]any{"metadata": map[string]any{"name": "test"}},
			expected: "metadata:\n  name: test\n",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "null\n",
		},
		{
			name:        "invalid yaml",
			input:       func() {}, // functions cannot be marshaled to YAML
			expected:    "invalid",
			expectError: true,
		},
		{
			name: "string with leading newline (block scalar normalization)",
			input: map[string]any{
				"data": map[string]any{
					"content": "\n{}\n",
				},
			},
			// The normalized output should NOT have a blank line after |
			expected: "data:\n  content: |\n    {}\n",
		},
		{
			name: "configmap with toJson nindent pattern",
			input: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"data": map[string]any{
					"overrides.json": "\n\n{}\n",
				},
			},
			// Should not have blank line after |
			expected: "apiVersion: v1\ndata:\n  overrides.json: |\n    {}\nkind: ConfigMap\n",
		},
		{
			name: "configmap with array toJson nindent pattern",
			input: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"data": []any{
					map[string]any{
						"overrides1.json": "\n{}\n",
					},
					map[string]any{
						"overrides2.json": "\n{}\n",
					},
				},
			},

			// Should not have blank line after |
			expected: "apiVersion: v1\ndata:\n  - overrides1.json: |\n      {}\n  - overrides2.json: |\n      {}\nkind: ConfigMap\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Defer a function to recover from panic
			defer func() {
				if r := recover(); r != nil {
					// We successfully recovered from panic
					assert.True(t, tt.expectError)
				}
			}()
			result := TrustedMarshalYAML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitBefore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{
			name:     "simple split",
			input:    "hello---world---test",
			sep:      "---",
			expected: []string{"hello", "---world", "---test"},
		},
		{
			name:     "starts with separator",
			input:    "---hello---world",
			sep:      "---",
			expected: []string{"", "---hello", "---world"},
		},
		{
			name:     "no separator",
			input:    "hello world",
			sep:      "---",
			expected: []string{"hello world"},
		},
		{
			name:     "empty string",
			input:    "",
			sep:      "---",
			expected: []string{""},
		},
		{
			name:     "separator only",
			input:    "---",
			sep:      "---",
			expected: []string{"", "---"},
		},
		{
			name:     "multiple consecutive separators",
			input:    "hello------world",
			sep:      "---",
			expected: []string{"hello", "-", "-", "-", "---world"},
		},
		{
			name:     "single character separator",
			input:    "a,b,c",
			sep:      ",",
			expected: []string{"a", ",b", ",c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitBefore(tt.input, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestYmlUnmarshalTestHelper(t *testing.T) {
	var result map[string]any

	// This should not fail
	YmlUnmarshalTestHelper("name: test", &result, t)
	assert.Equal(t, "test", result["name"])
}

func TestYmlMarshallTestHelper(t *testing.T) {
	input := map[string]any{"name": "test"}
	result := YmlMarshallTestHelper(input, t)
	assert.Contains(t, result, "name: test")
}
