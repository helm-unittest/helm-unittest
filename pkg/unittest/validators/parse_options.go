package validators

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
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

// ParseAware is satisfied only by validators that embed ParseOptions. It lets
// assertion parsing reject `parse` on assertions that do not support it,
// instead of silently ignoring the field.
//
// The unexported parseSpec method keeps this interface unimplementable from
// outside this package, so support cannot be claimed by accident.
type ParseAware interface {
	parseSpec() ParseOptions
	ValidateParseOptions() error
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
//
// An InnerPath starting with '.' is also rejected. wrapForPathLookup joins a
// non-object parse result's synthetic root key to InnerPath with "." when
// InnerPath doesn't start with "[", so a leading dot would produce a ".."
// expression. In JSONPath ".." means recursive descent, not child access, so
// it would silently match values at arbitrary depths instead of the intended
// child, rather than erroring or resolving to the intended child.
func (p ParseOptions) validate() error {
	if p.InnerPath != "" && strings.TrimSpace(p.InnerPath) == "" {
		return fmt.Errorf("field 'innerPath' must not be blank")
	}

	if strings.HasPrefix(strings.TrimSpace(p.InnerPath), ".") {
		return fmt.Errorf("field 'innerPath' must not start with '.'")
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

// ValidateParseOptions implements ParseAware.
func (p ParseOptions) ValidateParseOptions() error { return p.validate() }

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

// parsedRootKey is a synthetic map key used to wrap non-map parse results, so
// that innerPath can be resolved by the same path engine used for `path`.
// GetValueOfSetPath requires a map at the root; a parsed JSON array or scalar
// is not one.
const parsedRootKey = "__helmUnittestParsedRoot__"

// resolveParsed applies `parse` and `innerPath` to a single actual value.
//
// When parsing is not requested the value is returned unchanged, so callers can
// use this unconditionally.
//
// The result is a slice because innerPath, like path, may match several nodes
// (for example `servers[*].port`). Each match is validated independently by the
// caller's existing per-actual loop.
func (p ParseOptions) resolveParsed(actual any, path string) ([]any, error) {
	if !p.enabled() {
		return []any{actual}, nil
	}

	content, ok := actual.(string)
	if !ok {
		return nil, fmt.Errorf(
			"expect '%s' to be a string to parse as %s, got %s",
			path, p.Parse, describeType(actual),
		)
	}

	parsed, err := parseStructuredContent(content, p.Parse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse path '%s' as %s: %s", path, p.Parse, err)
	}

	return p.resolveInnerPath(parsed)
}

// resolveInnerPath applies InnerPath to an already-parsed value.
func (p ParseOptions) resolveInnerPath(parsed any) ([]any, error) {
	if p.InnerPath == "" {
		return []any{parsed}, nil
	}

	wrapper, effectivePath := wrapForPathLookup(parsed, p.InnerPath)

	return valueutils.GetValueOfSetPath(wrapper, effectivePath)
}

// wrapForPathLookup returns a map suitable for GetValueOfSetPath along with the
// path to use against it. A parsed object is used directly; anything else is
// nested under parsedRootKey and the path is prefixed accordingly.
func wrapForPathLookup(parsed any, innerPath string) (common.K8sManifest, string) {
	if asMap, ok := parsed.(map[string]any); ok {
		return asMap, innerPath
	}

	wrapper := common.K8sManifest{parsedRootKey: parsed}

	// An index expression attaches directly; a key expression needs a separator.
	if strings.HasPrefix(innerPath, "[") {
		return wrapper, parsedRootKey + innerPath
	}

	return wrapper, parsedRootKey + "." + innerPath
}

// describeType names a value's type for error messages, reporting untyped nil
// as "nil" rather than the empty string reflect returns.
func describeType(value any) string {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return "nil"
	}
	return valueType.String()
}
