package validators

import (
	"encoding/json"
	"fmt"
	"math"
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

// normalizeParsedNumbers converts json.Number values to int or float64 by
// integrality, not by written form, so that `parse: json` matches the type
// valueutils.GetValueOfSetPath produces for the exact same source value on an
// ordinary (non-parse) assertion path:
//
//	integral value      -> int      (e.g. 1, 1.0, 1e3, 0.0, -0.0)
//	non-integral value   -> float64 (e.g. 1.5, 0.1, 2.50)
//
// GetValueOfSetPath round-trips a value through a yaml.Node and decodes it
// back, which re-types by value rather than by lexical form. Typing by
// lexical form here would make `parse` disagree with every other assertion,
// and would make `parse` disagree with itself depending on whether innerPath
// is set, since innerPath resolution is routed through GetValueOfSetPath while
// a bare `parse` result is not.
//
// Using json.Decoder.UseNumber() keeps the original literal text available,
// which preserves the full precision of large integers that a direct float64
// decode would corrupt.
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

// maxInt64AsFloat is the smallest float64 that is out of int64's representable
// range (2^63; int64's actual max, 2^63-1, cannot itself be represented
// exactly as a float64). Used to guard against converting an integral float
// too large for int64 into a bogus, overflowed int.
const maxInt64AsFloat = 1 << 63

func normalizeJSONNumber(number json.Number) any {
	literal := number.String()

	// Integer-syntax literals are tried against Int64 first, before the float
	// path below, so that large integers keep exact int64 precision instead of
	// being rounded by a float64 round-trip.
	if !strings.ContainsAny(literal, ".eE") {
		if integer, err := number.Int64(); err == nil {
			return int(integer)
		}
	}

	float, err := number.Float64()
	if err != nil {
		// Unrepresentable as either int64 or float64; keep the literal so the
		// mismatch surfaces in the assertion failure rather than being
		// silently dropped.
		return literal
	}

	if float == math.Trunc(float) && float >= math.MinInt64 && float < maxInt64AsFloat {
		return int(int64(float))
	}

	return float
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

// resolveActuals applies parse and innerPath to every value matched by the
// assertion's path, flattening the results into a single slice.
//
// Flattening rather than nesting means a validator's existing per-actual loop
// works unchanged, and its value index increments across innerPath matches the
// same way it already does across path matches. When parsing is not requested
// the input is returned as-is, so callers can use this unconditionally.
func (p ParseOptions) resolveActuals(actuals []any, path string) ([]any, error) {
	if !p.enabled() {
		return actuals, nil
	}

	flattened := make([]any, 0, len(actuals))

	for _, actual := range actuals {
		parsed, err := p.resolveParsed(actual, path)
		if err != nil {
			return nil, err
		}

		flattened = append(flattened, parsed...)
	}

	return flattened, nil
}

// describePath names what an assertion looked up, for error messages. It
// includes the innerPath when one was used, so that two assertions probing
// different fields at the same innerPath produce distinguishable errors.
func (p ParseOptions) describePath(path string) string {
	if !p.enabled() || p.InnerPath == "" {
		return path
	}

	return fmt.Sprintf("%s innerPath %s", path, p.InnerPath)
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
