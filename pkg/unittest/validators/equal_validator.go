package validators

import (
	"reflect"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
)

// EqualValidator validate whether the value of Path equal to Value
type EqualValidator struct {
	Path  string
	Value interface{}
}

func (a EqualValidator) failInfo(actual interface{}, index int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(a.Value)
	customMessage := " to equal"
	if not {
		return splitInfof(
			setFailFormat(not, true, false, false, customMessage),
			index,
			a.Path,
			expectedYAML,
		)
	}

	actualYAML := common.TrustedMarshalYAML(actual)
	return splitInfof(
		setFailFormat(not, true, true, true, customMessage),
		index,
		a.Path,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

// Validate implement Validatable
func (a EqualValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		_, ok := actual.(string)
		if ok {
			actual = uniformContent(actual)
		}

		if reflect.DeepEqual(a.Value, actual) == context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo(actual, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
