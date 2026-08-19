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

const configJSON = `{"server":{"port":8080,"tls":{"enabled":true}},` +
	`"debug":false,"servers":[{"port":1},{"port":2}],` +
	`"panels":[{"title":"Latency"},{"title":"QPS"}]}`

func TestResolveParsedObjectRoot(t *testing.T) {
	tests := []struct {
		name      string
		innerPath string
		expected  []any
	}{
		{"whole document", "", []any{map[string]any{
			"server": map[string]any{
				"port": 8080,
				"tls":  map[string]any{"enabled": true},
			},
			"debug":   false,
			"servers": []any{map[string]any{"port": 1}, map[string]any{"port": 2}},
			"panels": []any{
				map[string]any{"title": "Latency"},
				map[string]any{"title": "QPS"},
			},
		}}},
		{"nested leaf", "server.port", []any{8080}},
		{"deeply nested leaf", "server.tls.enabled", []any{true}},
		{"boolean leaf", "debug", []any{false}},
		{"array element", "panels[0].title", []any{"Latency"}},
		{"array fan-out", "servers[*].port", []any{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := ParseOptions{Parse: ParseFormatJSON, InnerPath: tt.innerPath}
			actuals, err := options.resolveParsed(configJSON, "data[\"config.json\"]")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actuals)
		})
	}
}

func TestResolveParsedArrayRoot(t *testing.T) {
	const dashboardJSON = `[{"title":"Latency","n":1},{"title":"QPS","n":2}]`

	tests := []struct {
		name      string
		innerPath string
		expected  []any
	}{
		{"whole array", "", []any{[]any{
			map[string]any{"title": "Latency", "n": 1},
			map[string]any{"title": "QPS", "n": 2},
		}}},
		{"indexed element", "[0].title", []any{"Latency"}},
		{"indexed number", "[1].n", []any{2}},
		{"wildcard fan-out", "[*].title", []any{"Latency", "QPS"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := ParseOptions{Parse: ParseFormatJSON, InnerPath: tt.innerPath}
			actuals, err := options.resolveParsed(dashboardJSON, "data[\"d.json\"]")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actuals)
		})
	}
}

func TestResolveParsedYAMLContent(t *testing.T) {
	const configYAML = "server:\n  port: 8080\n  tls:\n    enabled: true\ndebug: false\n"

	options := ParseOptions{Parse: ParseFormatYAML, InnerPath: "server.port"}
	actuals, err := options.resolveParsed(configYAML, "data[\"config.yaml\"]")
	assert.NoError(t, err)
	assert.Equal(t, []any{8080}, actuals)
}

func TestResolveParsedDisabledPassesThrough(t *testing.T) {
	options := ParseOptions{}
	actuals, err := options.resolveParsed("plain string", "data.value")
	assert.NoError(t, err)
	assert.Equal(t, []any{"plain string"}, actuals)

	structured := map[string]any{"already": "parsed"}
	actuals, err = options.resolveParsed(structured, "data")
	assert.NoError(t, err)
	assert.Equal(t, []any{structured}, actuals)
}

func TestResolveParsedNonStringIsError(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON}

	_, err := options.resolveParsed(3, "spec.replicas")
	assert.EqualError(t, err,
		"expect 'spec.replicas' to be a string to parse as json, got int")

	_, err = options.resolveParsed(map[string]any{"a": 1}, "spec.selector")
	assert.EqualError(t, err,
		"expect 'spec.selector' to be a string to parse as json, got map[string]interface {}")

	_, err = options.resolveParsed(nil, "spec.missing")
	assert.EqualError(t, err,
		"expect 'spec.missing' to be a string to parse as json, got nil")
}

func TestResolveParsedParseFailureIncludesPath(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON}

	_, err := options.resolveParsed(`{"port": 8080,}`, `data["config.json"]`)
	assert.ErrorContains(t, err, `unable to parse path 'data["config.json"]' as json:`)
}

// An innerPath that matches nothing yields an empty slice with a nil error.
// Each validator's existing len(actuals) == 0 branch then handles it, which
// keeps negative (not:) assertions behaving consistently.
func TestResolveParsedUnmatchedInnerPathIsEmpty(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: "nope.deeper"}

	actuals, err := options.resolveParsed(`{"a":1}`, "data.config")
	assert.NoError(t, err)
	assert.Empty(t, actuals)
}

func TestResolveParsedInvalidInnerPathIsError(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: "a[["}

	_, err := options.resolveParsed(`{"a":1}`, "data.config")
	assert.Error(t, err)
}
