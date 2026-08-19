package validators

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

// Supported values for the `parse` assertion field.
const (
	ParseFormatJSON = "json"
	ParseFormatYAML = "yaml"
)

// ParseOptions is embedded by validators that support parsing the string value
// at Path as structured content before asserting on it.
//
// Validators embed it with `mapstructure:",squash"` so that `parse` and
// `innerPath` remain top-level keys in the test suite YAML, consistent with
// the existing `decodeBase64` field.
type ParseOptions struct {
	Parse     string `yaml:"parse"`
	InnerPath string `yaml:"innerPath"`
}

// parseSpec exposes the options via method promotion, so that any validator
// embedding ParseOptions satisfies parseAware.
func (p ParseOptions) parseSpec() ParseOptions { return p }

// parseAware is satisfied only by validators that embed ParseOptions. It lets
// assertion parsing reject `parse` on assertions that do not support it,
// instead of silently ignoring the field.
type parseAware interface {
	parseSpec() ParseOptions
}

// validate checks the option combination is coherent, independent of any
// actual value.
//
// An InnerPath consisting solely of whitespace is rejected up front, before
// the other checks: it is indistinguishable from a typo, and the underlying
// path engine's behavior for such input is inconsistent (some whitespace
// produces a cryptic parse error, other whitespace produces no matches and
// no error at all, which downstream validators cannot tell apart from an
// absent value). An empty InnerPath remains valid; it means "use the whole
// parsed document".
func (p ParseOptions) validate() error {
	if p.InnerPath != "" && strings.TrimSpace(p.InnerPath) == "" {
		return fmt.Errorf("field 'innerPath' must not be blank")
	}

	if p.Parse == "" {
		if p.InnerPath != "" {
			return fmt.Errorf("field 'innerPath' requires 'parse' to be set")
		}
		return nil
	}

	if p.Parse != ParseFormatJSON && p.Parse != ParseFormatYAML {
		return fmt.Errorf(
			"invalid parse format '%s', expected '%s' or '%s'",
			p.Parse, ParseFormatJSON, ParseFormatYAML,
		)
	}

	return nil
}

// enabled reports whether parsing was requested.
func (p ParseOptions) enabled() bool { return p.Parse != "" }

// normalizeParsedNumbers converts json.Number values to int or float64 using
// the same rule go.yaml.in/yaml/v3 applies, so that `parse: json` and
// `parse: yaml` produce identical Go types for identical input:
//
//	integer syntax                     -> int
//	decimal point or exponent syntax   -> float64
//
// Using json.Decoder.UseNumber() keeps the original literal text available for
// this decision, and preserves the full precision of large integers that a
// direct float64 decode would corrupt.
func normalizeParsedNumbers(value any) any {
	switch typed := value.(type) {
	case json.Number:
		return normalizeJSONNumber(typed)
	case map[string]any:
		normalized := make(map[string]any, len(typed))
		for key, element := range typed {
			normalized[key] = normalizeParsedNumbers(element)
		}
		return normalized
	case []any:
		normalized := make([]any, len(typed))
		for index, element := range typed {
			normalized[index] = normalizeParsedNumbers(element)
		}
		return normalized
	default:
		return value
	}
}

func normalizeJSONNumber(number json.Number) any {
	literal := number.String()

	if !strings.ContainsAny(literal, ".eE") {
		if integer, err := number.Int64(); err == nil {
			return int(integer)
		}
	}

	if float, err := number.Float64(); err == nil {
		return float
	}

	// Unrepresentable as either; keep the literal so the mismatch surfaces in
	// the assertion failure rather than being silently dropped.
	return literal
}

// parseStructuredContent parses content in the given format and normalizes
// numbers so both formats agree on types.
func parseStructuredContent(content, format string) (any, error) {
	switch format {
	case ParseFormatJSON:
		decoder := json.NewDecoder(strings.NewReader(content))
		decoder.UseNumber()

		var parsed any
		if err := decoder.Decode(&parsed); err != nil {
			return nil, err
		}
		if decoder.More() {
			return nil, fmt.Errorf("unexpected trailing content after JSON value")
		}
		return normalizeParsedNumbers(parsed), nil

	case ParseFormatYAML:
		var parsed any
		if err := common.YmlUnmarshal(content, &parsed); err != nil {
			return nil, err
		}
		return parsed, nil

	default:
		return nil, fmt.Errorf(
			"invalid parse format '%s', expected '%s' or '%s'",
			format, ParseFormatJSON, ParseFormatYAML,
		)
	}
}
