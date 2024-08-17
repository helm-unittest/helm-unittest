package validators

import (
	"fmt"
	"reflect"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// operatorValidator validate whether the value of Path is according to the behaviour of operatorgreater to Value
// internal struct, not exposed to end user
type operatorValidator struct {
	Path           string
	Value          interface{}
	ComparisonType string
}

func (o operatorValidator) failInfo(msg, comparisonType string, index int, not bool) []string {
	customMsg := fmt.Sprintf(" to be %s then or equal to, got", comparisonType)
	return splitInfof(
		setFailFormat(not, true, false, false, customMsg),
		index,
		o.Path,
		msg,
	)
}

// compareValues performs a validation of a Kubernetes manifest against an expected value.
// It compares the actual value retrieved from the manifest with the expected value,
// ensuring that they are of compatible types and that the actual value is greater, less than or equal to the expected value.
// If successful, it returns true along with a nil error slice. If unsuccessful, it returns false
// along with an error slice containing information about the validation failure.
func (o operatorValidator) compareValues(expected, actual interface{}, comparisonType string, negative bool) (bool, []string) {
	expStr := fmt.Sprintf("%v", expected)
	actStr := fmt.Sprintf("%v", actual)
	result := false

	switch exp := expected.(type) {
	case string:
		result = o.compareStringValues(exp, actual.(string), comparisonType, negative)
	case int:
		result = o.compareIntValues(exp, actual.(int), comparisonType, negative)
	case float64:
		result = o.compareFloatValues(exp, actual.(float64), comparisonType, negative)
	default:
		return false, []string{fmt.Sprintf("unsupported type '%T'", expected)}
	}

	if !result {
		return false, []string{fmt.Sprintf("the expected '%s' is not %s or equal to the actual '%s'", expStr, comparisonType, actStr)}
	}

	return true, nil
}

func (o operatorValidator) compareStringValues(expected, actual string, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && expected >= actual) || (comparisonType == "less" && expected <= actual) == negative {
		return true
	}
	return false
}

func (o operatorValidator) compareIntValues(expected, actual int, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && expected >= actual) || (comparisonType == "less" && expected <= actual) == negative {
		return true
	}
	return false
}

func (o operatorValidator) compareFloatValues(expected, actual float64, comparisonType string, negative bool) bool {
	if (comparisonType == "greater" && expected >= actual) || (comparisonType == "less" && expected <= actual) == negative {
		return true
	}
	return false
}

// Validate implement Validatable
func (o operatorValidator) Validate(context *ValidateContext) (bool, []string) {
	log.WithField("validator", o.ComparisonType).Debugln("expected content:", o.Value, "path:", o.Path)
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, o.Path)
		if err != nil {
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if len(actual) == 0 {
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown path '%s'", o.Path))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		actType := reflect.TypeOf(actual[0])
		expType := reflect.TypeOf(o.Value)

		if actType != expType {
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("actual '%s' and expected '%s' types do not match", actType, expType))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		result, errors := o.compareValues(o.Value, actual[0], o.ComparisonType, !context.Negative)
		if errors != nil {
			errorMessage := o.failInfo(errors[0], o.ComparisonType, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		}

		validateSuccess = determineSuccess(idx, validateSuccess, result)
	}

	return validateSuccess, validateErrors
}
