package validators

import (
	"cmp"
	"fmt"
	"reflect"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

// FailedTemplateValidator validate whether the errorMessage equal to errorMessage
type FailedTemplateValidator struct {
	ErrorMessage string
	ErrorPattern string
}

func (a FailedTemplateValidator) failInfo(actual interface{}, index int, not bool) []string {
	customMessage := " to equal"
	if a.ErrorPattern != "" {
		customMessage = " to match"
	} else if a.ErrorPattern + a.ErrorMessage == "" {
		customMessage = " to throw"
	}

	message := cmp.Or(a.ErrorMessage, a.ErrorPattern, cmp.Or(fmt.Sprintf("%s", actual), "error"))

	log.WithField("validator", "failed_template").Debugln("expected content:", message)
	log.WithField("validator", "failed_template").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			index,
			message,
		)
	}

	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		index,
		message,
		fmt.Sprintf("%s", actual),
	)
}

// Validate implement Validatable
func (a FailedTemplateValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	if a.ErrorMessage != "" && a.ErrorPattern != "" {
		errorMessage := splitInfof(errorFormat, -1, "single attribute 'errorMessage' or 'errorPattern' supported at the same time")
		validateErrors = append(validateErrors, errorMessage...)
	} else if context.RenderError != nil {
		if a.ErrorMessage != "" && reflect.DeepEqual(a.ErrorMessage, context.RenderError.Error()) == context.Negative && a != (FailedTemplateValidator{}) {
			errorMessage := a.failInfo(context.RenderError.Error(), -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		} else {
			validateSuccess = true
		}
	} else {
		validateSuccess, validateErrors = a.validateManifests(manifests, context)
	}

	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateManifests(manifests []common.K8sManifest, context *ValidateContext) (bool, []string) {
	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		currentSuccess := false
		actual := manifest[common.RAW]


		if a == (FailedTemplateValidator{}) && !context.Negative {
			// If the validator is empty and the context is negative,
			// continue to the next iteration without throwing an error.
			continue
		}

		if a.ErrorPattern != "" {
			currentSuccess, validateErrors = a.validateErrorPattern(actual, idx, context, currentSuccess, validateErrors)
		} else if a.ErrorMessage != "" {
			currentSuccess, validateErrors = a.validateErrorMessage(actual, idx, context, currentSuccess, validateErrors)
		} else {
			currentSuccess, validateErrors = a.validateNoError(actual, idx, context, currentSuccess, validateErrors)
		}

		validateSuccess = determineSuccess(idx, validateSuccess, currentSuccess)
	}

	if len(manifests) == 0 && !context.Negative {
		validateSuccess = false
		errorMessage := a.failInfo("No failed document", -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateErrorPattern(actual interface{}, idx int, context *ValidateContext, validateSuccess bool, validateErrors []string) (bool, []string) {
	p, err := regexp.Compile(a.ErrorPattern)
	if err != nil {
		errorMessage := splitInfof(errorFormat, -1, err.Error())
		validateErrors = append(validateErrors, errorMessage...)
		return validateSuccess, validateErrors
	}

	if p.MatchString(actual.(string)) == context.Negative {
		errorMessage := a.failInfo(actual, idx, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateErrorMessage(actual interface{}, idx int, context *ValidateContext, validateSuccess bool, validateErrors []string) (bool, []string) {
	if reflect.DeepEqual(a.ErrorMessage, actual.(string)) == context.Negative {
		errorMessage := a.failInfo(actual, idx, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else {
		validateSuccess = true
	}
	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateNoError(actual interface{}, idx int, context *ValidateContext, validateSuccess bool, validateErrors []string) (bool, []string) {
	if actual != nil && context.Negative {
		errorMessage := a.failInfo(actual, idx, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if actual == nil && context.Negative {
		validateSuccess = true
	}
	return validateSuccess, validateErrors
}
