package validators

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
)

// Numbers are typed by value, not by written form, matching what
// valueutils.GetValueOfSetPath produces for an ordinary path: an integral value
// becomes int, a non-integral value stays float64. This keeps parse consistent
// with every other assertion, and consistent with itself whether or not
// innerPath is used (innerPath resolution round-trips through the path engine,
// which applies the same rule).
func TestNormalizeParsedNumbersByValue(t *testing.T) {
	tests := []struct {
		name     string
		literal  string
		expected any
	}{
		{"integer", "1", 1},
		{"large integer", "8080", 8080},
		{"zero", "0", 0},
		{"integral with decimal point", "1.0", 1},
		{"integral with trailing zero", "2.50", 2.5},
		{"integral exponent", "1e3", 1000},
		{"integral exponent with decimal", "1.0e2", 100},
		{"negative zero", "-0.0", 0},
		{"fractional", "1.5", 1.5},
		{"small fractional", "0.1", 0.1},
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

// The parse rule must agree with what the assertion path engine produces, so
// that `parse` behaves like every other assertion and so that parse with and
// without innerPath agree with each other.
func TestNormalizedNumbersMatchPathEngine(t *testing.T) {
	literals := []string{"1", "1.0", "1.5", "1e3", "1.0e2", "0.0", "-0.0", "2.50", "8080", "0.1"}

	for _, literal := range literals {
		t.Run(literal, func(t *testing.T) {
			fromParse, err := parseStructuredContent(`{"a":`+literal+`}`, ParseFormatJSON)
			assert.NoError(t, err)
			parsedValue := fromParse.(map[string]any)["a"]

			var manifest common.K8sManifest
			assert.NoError(t, common.YmlUnmarshal("a: "+literal, &manifest))
			fromPath, err := valueutils.GetValueOfSetPath(manifest, "a")
			assert.NoError(t, err)
			assert.Len(t, fromPath, 1)

			assert.IsType(t, fromPath[0], parsedValue,
				"parse and the path engine must agree on the type of %s", literal)
			assert.Equal(t, fromPath[0], parsedValue,
				"parse and the path engine must agree on the value of %s", literal)
		})
	}
}

// parse: json and parse: yaml must agree on types for every literal, including
// integral floats and exponent forms, since both now normalize numbers by
// value rather than by written form.
func TestParseJSONAndYAMLAgreeOnTypes(t *testing.T) {
	literals := []string{
		"1", "8080", "0", "1.5", "9007199254740993",
		"1.0", "1e3", "8080.0", "1.0e2", "0.0",
	}

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
	assert.Equal(t, 2, list[1], "2.0 is integral, so it normalizes to int")
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
		{
			name:        "innerPath with a leading dot is rejected",
			options:     ParseOptions{Parse: ParseFormatJSON, InnerPath: ".server.port"},
			expectedErr: "field 'innerPath' must not start with '.'",
		},
		{
			name:        "innerPath that is only a dot is rejected",
			options:     ParseOptions{Parse: ParseFormatJSON, InnerPath: "."},
			expectedErr: "field 'innerPath' must not start with '.'",
		},
		{
			name:        "innerPath with a leading dot before an index is rejected",
			options:     ParseOptions{Parse: ParseFormatJSON, InnerPath: ".[0]"},
			expectedErr: "field 'innerPath' must not start with '.'",
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

// A leading dot must be rejected rather than silently joined to the synthetic
// root key, which would produce a '..' recursive-descent expression and match
// values at arbitrary depths instead of the requested child.
func TestParseOptionsRejectsLeadingDotBeforeResolution(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: ".x"}

	assert.EqualError(t, options.validate(),
		"field 'innerPath' must not start with '.'")
}

func TestResolveActualsFlattensFanOut(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"}

	actuals := []any{`{"servers":[{"port":1},{"port":2}]}`}

	flattened, err := options.resolveActuals(actuals, "data.config")

	assert.NoError(t, err)
	assert.Equal(t, []any{1, 2}, flattened,
		"innerPath matches must be flattened into the actuals slice")
}

func TestResolveActualsFlattensAcrossMultipleActuals(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"}

	actuals := []any{`{"port":1}`, `{"port":2}`, `{"port":3}`}

	flattened, err := options.resolveActuals(actuals, "data.*")

	assert.NoError(t, err)
	assert.Equal(t, []any{1, 2, 3}, flattened,
		"each path match contributes its own innerPath matches, in order")
}

func TestResolveActualsDisabledReturnsInputUnchanged(t *testing.T) {
	options := ParseOptions{}

	actuals := []any{"a", 2, map[string]any{"b": 3}}

	flattened, err := options.resolveActuals(actuals, "data.config")

	assert.NoError(t, err)
	assert.Equal(t, actuals, flattened)
}

func TestResolveActualsPropagatesError(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON}

	_, err := options.resolveActuals([]any{`{"bad":}`}, "data.config")

	assert.ErrorContains(t, err, "unable to parse path 'data.config' as json")
}

func TestResolveActualsUnmatchedInnerPathYieldsEmpty(t *testing.T) {
	options := ParseOptions{Parse: ParseFormatJSON, InnerPath: "nope"}

	flattened, err := options.resolveActuals([]any{`{"a":1}`}, "data.config")

	assert.NoError(t, err)
	assert.Empty(t, flattened)
}

func TestDescribePathNamesBothPaths(t *testing.T) {
	assert.Equal(t, "data.config",
		ParseOptions{}.describePath("data.config"))

	assert.Equal(t, "data.config",
		ParseOptions{Parse: ParseFormatJSON}.describePath("data.config"))

	assert.Equal(t, "data.config innerPath server.port",
		ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"}.describePath("data.config"))
}
