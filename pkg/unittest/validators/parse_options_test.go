package validators

import (
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
