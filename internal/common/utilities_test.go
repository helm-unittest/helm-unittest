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

func TestYmlUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "valid yaml to map",
			input: "name: test\nversion: 1.0",
		},
		{
			name:  "empty yaml",
			input: "",
		},
		{
			name:        "invalid yaml",
			input:       "invalid:\n  - yaml\n- structure",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]any
			err := YmlUnmarshal(tt.input, &result)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestYmlMarshall(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := YmlMarshall(tt.input)
			assert.NoError(t, err)
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
