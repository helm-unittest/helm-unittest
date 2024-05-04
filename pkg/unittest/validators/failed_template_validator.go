package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

// FailedTemplateValidator validate whether the errorMessage equal to errorMessage
type FailedTemplateValidator struct {
	ErrorMessage string
}

func (a FailedTemplateValidator) failInfo(actual interface{}, index int, not bool) []string {
	customMessage := " to equal"

	log.WithField("validator", "failed_template").Debugln("expected content:", a.ErrorMessage)
	log.WithField("validator", "failed_template").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			index,
			a.ErrorMessage,
		)
	}

	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		index,
		a.ErrorMessage,
		fmt.Sprintf("%s", actual),
	)
}

// Validate implement Validatable
func (a FailedTemplateValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	if context.RenderError != nil {

		if reflect.DeepEqual(a.ErrorMessage, context.RenderError.Error()) == context.Negative && a != (FailedTemplateValidator{}) {
			validateSuccess = false
			errorMessage := a.failInfo(context.RenderError.Error(), -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		}

	} else {

		for idx, manifest := range manifests {
			actual := manifest[common.RAW]

			if a == (FailedTemplateValidator{}) && !context.Negative {
				// If the validator is empty and the context is not negative,
				// continue to the next iteration without throwing an error.
				continue
			}
			if reflect.DeepEqual(a.ErrorMessage, actual) == context.Negative {
				validateSuccess = false
				errorMessage := a.failInfo(actual, idx, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}
			validateSuccess = determineSuccess(idx, validateSuccess, true)
		}

		if len(manifests) == 0 && !context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo("No failed document", -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		}
	}

	return validateSuccess, validateErrors
}
