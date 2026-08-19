package validators

import (
	"encoding/base64"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// EqualValidator validate whether the value of Path equal to Value
type EqualValidator struct {
	Path         string
	Value        any
	DecodeBase64 bool `yaml:"decodeBase64"`
	ParseOptions `mapstructure:",squash"`
}

func (a EqualValidator) failInfo(actual any, manifestIndex, actualIndex int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(a.Value)
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to equal"

	log.WithField("validator", "equal").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "equal").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, true, false, false, customMessage),
			manifestIndex,
			actualIndex,
			a.Path,
			expectedYAML,
		)
	}

	return splitInfof(
		setFailFormat(not, true, true, true, customMessage),
		manifestIndex,
		actualIndex,
		a.Path,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

func (a EqualValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	actuals, decodeErrors := a.decodeBase64Actuals(actuals, manifestIndex)
	if decodeErrors != nil {
		return false, decodeErrors
	}

	actuals, err = a.ParseOptions.resolveActuals(actuals, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1,
			fmt.Sprintf("unknown path %s", a.ParseOptions.describePath(a.Path)))
	}

	validateManifestSuccess := (len(actuals) == 0 && context.Negative)
	var validateManifestErrors []string

	for actualIndex, actual := range actuals {
		validateSingleSuccess, validateSingleErrors := a.validateSingleActual(actual, manifestIndex, actualIndex, context)
		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)
		if !validateSingleSuccess && context.FailFast {
			break
		}
	}

	return validateManifestSuccess, validateManifestErrors
}

// decodeBase64Actuals decodes each string actual, so that parsing operates on
// the decoded content. Ordering matters: a Secret can hold base64-encoded
// JSON/YAML, so decoding must happen before `parse` and `innerPath` are
// applied. Non-string actuals pass through untouched, matching the previous
// per-actual behavior.
//
// The returned []string is non-nil only on failure, in which case it is the
// complete splitInfof error output and the caller must return it immediately.
func (a EqualValidator) decodeBase64Actuals(actuals []any, manifestIndex int) ([]any, []string) {
	if !a.DecodeBase64 {
		return actuals, nil
	}

	decoded := make([]any, len(actuals))

	for actualIndex, actual := range actuals {
		s, ok := actual.(string)
		if !ok {
			decoded[actualIndex] = actual
			continue
		}

		decodedSingleActual, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("unable to decode base64 expected content %s", actual))
		}

		decoded[actualIndex] = string(decodedSingleActual)
	}

	return decoded, nil
}

func (a EqualValidator) validateSingleActual(actual any, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	return a.compareActual(actual, manifestIndex, actualIndex, context)
}

// compareActual performs the equality comparison against a single value.
func (a EqualValidator) compareActual(actual any, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	normalized := normalizeActual(actual)

	if reflect.DeepEqual(a.Value, normalized) == context.Negative {
		return false, a.failInfo(normalized, manifestIndex, actualIndex, context.Negative)
	}

	return true, []string{}
}

// normalizeActual recursively applies uniformContent to all string values
// within maps and slices, so that leading newlines and trailing spaces before
// newlines don't cause DeepEqual to fail on semantically identical content.
func normalizeActual(v any) any {
	switch val := v.(type) {
	case string:
		return uniformContent(val)
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			result[k] = normalizeActual(v)
		}
		return result
	case map[any]any:
		result := make(map[any]any, len(val))
		for k, v := range val {
			result[k] = normalizeActual(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = normalizeActual(v)
		}
		return result
	default:
		return v
	}
}

// Validate implement Validatable
func (a EqualValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		validateManifestSuccess, validateManifestErrors := a.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, validateManifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := a.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
