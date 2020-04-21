package validators

import (
	"reflect"

	"github.com/lrills/helm-unittest/unittest/common"
)

// EqualRawValidator validate whether the raw value equal to Value
type EqualRawValidator struct {
	Value string
}

func (a EqualRawValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to equal"
	}
	failFormat := `
Expected` + notAnnotation + `:
%s`

	expectedYAML := common.TrustedMarshalYAML(a.Value)
	if not {
		return splitInfof(failFormat, -1, expectedYAML)
	}

	actualYAML := common.TrustedMarshalYAML(actual)
	return splitInfof(
		failFormat+`
Actual:
%s
Diff:
%s
`,
		-1,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

// Validate implement Validatable
func (a EqualRawValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		if reflect.DeepEqual(a.Value, actual) == context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo(actual, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
