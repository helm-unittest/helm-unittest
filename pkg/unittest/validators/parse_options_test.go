package validators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeParsedNumbersByLexicalForm(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected any
	}{
		{"integer", "1", 1},
		{"large integer", "8080", 8080},
		{"zero", "0", 0},
		{"integral with decimal point stays float", "1.0", float64(1)},
		{"fractional", "1.5", 1.5},
		{"exponent", "1e3", float64(1000)},
		{"negative zero float", "-0.0", float64(0)},
		{"int64 precision preserved", "9007199254740993", 9007199254740993},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseStructuredContent(`{"a":`+tt.literal+`}`, ParseFormatJSON)
			assert.NoError(t, err)
			actual := parsed.(map[string]any)["a"]
			assert.Equal(t, tt.expected, actual)
			assert.IsType(t, tt.expected, actual)
		})
	}
}

func TestParseJSONAndYAMLAgreeOnTypes(t *testing.T) {
	literals := []string{"1", "8080", "0", "1.0", "1.5", "1e3", "9007199254740993"}

	for _, literal := range literals {
		t.Run(literal, func(t *testing.T) {
			fromJSON, err := parseStructuredContent(`{"a":`+literal+`}`, ParseFormatJSON)
			assert.NoError(t, err)
			fromYAML, err := parseStructuredContent("a: "+literal, ParseFormatYAML)
			assert.NoError(t, err)

			assert.Equal(t, fromYAML, fromJSON,
				"parse: json and parse: yaml must produce identical values for %s", literal)
			assert.IsType(t,
				fromYAML.(map[string]any)["a"],
				fromJSON.(map[string]any)["a"],
				"parse: json and parse: yaml must produce identical types for %s", literal)
		})
	}
}

func TestParseStructuredContentNestedNormalization(t *testing.T) {
	parsed, err := parseStructuredContent(
		`{"outer":{"n":1,"f":1.5},"list":[1,2.0,{"deep":3}]}`, ParseFormatJSON)
	assert.NoError(t, err)

	root := parsed.(map[string]any)
	outer := root["outer"].(map[string]any)
	assert.Equal(t, 1, outer["n"])
	assert.Equal(t, 1.5, outer["f"])

	list := root["list"].([]any)
	assert.Equal(t, 1, list[0])
	assert.Equal(t, float64(2), list[1])
	assert.Equal(t, 3, list[2].(map[string]any)["deep"])
}

func TestParseOptionsValidate(t *testing.T) {
	tests := []struct {
		name        string
		options     ParseOptions
		expectedErr string
	}{
		{
			name:    "empty is valid",
			options: ParseOptions{},
		},
		{
			name:    "json is valid",
			options: ParseOptions{Parse: ParseFormatJSON},
		},
		{
			name:    "yaml is valid",
			options: ParseOptions{Parse: ParseFormatYAML},
		},
		{
			name:    "json with innerPath is valid",
			options: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"},
		},
		{
			name:        "unknown format is rejected",
			options:     ParseOptions{Parse: "toml"},
			expectedErr: "invalid parse format 'toml', expected 'json' or 'yaml'",
		},
		{
			name:        "innerPath without parse is rejected",
			options:     ParseOptions{InnerPath: "server.port"},
			expectedErr: "field 'innerPath' requires 'parse' to be set",
		},
		{
			name:        "whitespace-only innerPath is rejected",
			options:     ParseOptions{Parse: ParseFormatJSON, InnerPath: "   "},
			expectedErr: "field 'innerPath' must not be blank",
		},
		{
			name:        "tab-only innerPath is rejected",
			options:     ParseOptions{Parse: ParseFormatJSON, InnerPath: "\t"},
			expectedErr: "field 'innerPath' must not be blank",
		},
		{
			name:        "whitespace-only innerPath without parse is rejected",
			options:     ParseOptions{InnerPath: "   "},
			expectedErr: "field 'innerPath' must not be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.validate()
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.expectedErr)
		})
	}
}

func TestParseStructuredContentErrors(t *testing.T) {
	_, err := parseStructuredContent(`{"port": 8080,}`, ParseFormatJSON)
	assert.Error(t, err, "malformed JSON must be rejected by the json parser")

	_, err = parseStructuredContent(`{"a":1} {"b":2}`, ParseFormatJSON)
	assert.EqualError(t, err, "unexpected trailing content after JSON value")

	_, err = parseStructuredContent("a: [unclosed", ParseFormatYAML)
	assert.Error(t, err, "malformed YAML must be rejected")

	_, err = parseStructuredContent(`{}`, "toml")
	assert.EqualError(t, err, "invalid parse format 'toml', expected 'json' or 'yaml'")
}

// A number representable as neither int64 nor float64 falls back to its literal
// string, matching what go.yaml.in/yaml/v3 does for the same input. Returning an
// error instead would break the guarantee that json and yaml agree on types.
func TestParseStructuredContentOversizedNumberFallsBackToLiteral(t *testing.T) {
	oversized := strings.Repeat("9", 400)

	fromJSON, err := parseStructuredContent(`{"a":`+oversized+`}`, ParseFormatJSON)
	assert.NoError(t, err)

	fromYAML, err := parseStructuredContent("a: "+oversized, ParseFormatYAML)
	assert.NoError(t, err)

	assert.Equal(t, oversized, fromJSON.(map[string]any)["a"])
	assert.IsType(t,
		fromYAML.(map[string]any)["a"],
		fromJSON.(map[string]any)["a"],
		"json and yaml must agree on type even for unrepresentable numbers")
}
