package validators

import (
	"fmt"
	"reflect"

	"github.com/lrills/helm-unittest/unittest/common"
)

// FailedTemplateValidator validate whether the errorMessage equal to errorMessage
type FailedTemplateValidator struct {
	ErrorMessage string
}

func (a FailedTemplateValidator) failInfo(actual interface{}, index int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to equal"
	}
	failFormat := `
Expected` + notAnnotation + `:
%s`

	if not {
		return splitInfof(failFormat, index, a.ErrorMessage)
	}

	return splitInfof(
		failFormat+`
Actual:
%s
`,
		index,
		a.ErrorMessage,
		fmt.Sprintf("%v", actual),
	)
}

// Validate implement Validatable
func (a FailedTemplateValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual := manifest[common.RAW]

		if reflect.DeepEqual(a.ErrorMessage, actual) == context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo(actual, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if idx == 0 {
			validateSuccess = true
		}

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
