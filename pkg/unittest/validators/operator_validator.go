package validators

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// operatorValidator validate whether the value of Path is according to the behaviour of operatorgreater to Value
// internal struct, not exposed to end user
type operatorValidator struct {
	Path           string
	Value          any
	ComparisonType string
}

func (o operatorValidator) failInfo(msg, comparisonType string, manifestIndex, actualIndex int, not bool) []string {
	customMsg := fmt.Sprintf(" to be %s then or equal to, got", comparisonType)
	return splitInfof(
		setFailFormat(not, true, false, false, customMsg),
		manifestIndex,
		actualIndex,
		o.Path,
		msg,
	)
}

// compareValues performs a validation of a Kubernetes manifest against an expected value.
// It compares the actual value retrieved from the manifest with the expected value,
// ensuring that they are of compatible types and that the actual value is greater, less than or equal to the expected value.
// If successful, it returns true along with a nil error slice. If unsuccessful, it returns false
// along with an error slice containing information about the validation failure.
func (o operatorValidator) compareValues(expected, actual any, comparisonType string, negative bool) (bool, []string) {
	expStr := fmt.Sprintf("%v", expected)
	actStr := fmt.Sprintf("%v", actual)
	result := false

	switch exp := expected.(type) {
	case string:
		// Try Kubernetes resource quantity parsing first (e.g., "50m", "256Mi", "1Gi")
		if actStr, ok := actual.(string); ok {
			if expQ, expErr := parseK8sResourceQuantity(exp); expErr == nil {
				if actQ, actErr := parseK8sResourceQuantity(actStr); actErr == nil {
					result = compareResourceQuantities(expQ, actQ, comparisonType, negative)
					if result {
						return true, nil
					}
					return false, []string{fmt.Sprintf("the actual '%s' is not %s or equal to the expected '%s'", actStr, comparisonType, expStr)}
				}
			}
		}
		result = o.compareStringValues(exp, actStr, comparisonType, negative)
	case int:
		result = o.compareIntValues(exp, actual.(int), comparisonType, negative)
	case float64:
		result = o.compareFloatValues(exp, actual.(float64), comparisonType, negative)
	default:
		return false, []string{fmt.Sprintf("unsupported type '%T'", expected)}
	}

	if !result {
		return false, []string{fmt.Sprintf("the actual '%s' is not %s or equal to the expected '%s'", actStr, comparisonType, expStr)}
	}

	return true, nil
}

// parseK8sResourceQuantity parses a Kubernetes resource string like "50m", "256Mi", "1Gi"
// Returns the quantity in base units (millicores for CPU, bytes for memory) or an error if not parseable
func parseK8sResourceQuantity(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// Try plain number first
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v, nil
	}

	// Parse Kubernetes resource suffixes
	multipliers := map[string]float64{
		"m":  0.001,  // millicores
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"Pi": 1024 * 1024 * 1024 * 1024 * 1024,
		"Ei": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
		"P":  1000 * 1000 * 1000 * 1000 * 1000,
		"E":  1000 * 1000 * 1000 * 1000 * 1000 * 1000,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numPart := s[:len(s)-len(suffix)]
			if v, err := strconv.ParseFloat(numPart, 64); err == nil {
				return v * mult, nil
			}
		}
	}

	return 0, fmt.Errorf("not a valid Kubernetes resource quantity: %s", s)
}

// compareResourceQuantities compares two parsed quantities
func compareResourceQuantities(expected, actual float64, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && actual >= expected) || (comparisonType == "less" && actual <= expected) == negative {
		return true
	}
	return false
}

func (o operatorValidator) compareStringValues(expected, actual string, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && actual >= expected) || (comparisonType == "less" && actual <= expected) == negative {
		return true
	}
	return false
}

func (o operatorValidator) compareIntValues(expected, actual int, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && actual >= expected) || (comparisonType == "less" && actual <= expected) == negative {
		return true
	}
	return false
}

func (o operatorValidator) compareFloatValues(expected, actual float64, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && actual >= expected) || (comparisonType == "less" && actual <= expected) == negative {
		return true
	}
	return false
}

func (o operatorValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, o.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path '%s'", o.Path))
	}

	validateManifestSuccess := (len(actuals) == 0 && context.Negative)
	var validateManifestErrors []string

	for actualIndex, actual := range actuals {
		actType := reflect.TypeOf(actual)
		expType := reflect.TypeOf(o.Value)

		validateSingleSuccess := false
		var validateSingleErrors []string

		if actType != expType {
			errorMessage := splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("actual '%s' and expected '%s' types do not match", actType, expType))
			validateManifestErrors = append(validateManifestErrors, errorMessage...)
			continue
		}

		validateSingleSuccess, errors := o.compareValues(o.Value, actual, o.ComparisonType, !context.Negative)
		if errors != nil {
			errorMessage := o.failInfo(errors[0], o.ComparisonType, manifestIndex, actualIndex, context.Negative)
			validateSingleErrors = append(validateSingleErrors, errorMessage...)
		}

		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)

		if !validateManifestSuccess && context.FailFast {
			break
		}
	}

	return validateManifestSuccess, validateManifestErrors
}

// Validate implement Validatable
func (o operatorValidator) Validate(context *ValidateContext) (bool, []string) {
	log.WithField("validator", o.ComparisonType).Debugln("expected content:", o.Value, "path:", o.Path)
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		validateManifestSuccess, validateManifestErrors := o.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, validateManifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := o.failInfo("no manifests found", o.ComparisonType, -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
