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
	}

	message := cmp.Or(a.ErrorMessage, a.ErrorPattern)

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
		return false, validateErrors
	}

	if context.RenderError != nil {
		// Validating error, when the errorSource is due to rendering errors
		if a.ErrorPattern != "" {
			return a.validateErrorPattern(context.RenderError.Error(), -1, context)
		} else if a.ErrorMessage != "" {
			return a.validateErrorMessage(context.RenderError.Error(), -1, context)
		} else {
			validateSuccess = true
		}
	} else {
		var errorsToAppend []string
		validateSuccess, errorsToAppend = a.validateManifests(manifests, context)
		validateErrors = append(validateErrors, errorsToAppend...)
	}

	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateManifests(manifests []common.K8sManifest, context *ValidateContext) (bool, []string) {
	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		currentSuccess := false
		validateSingleErrors := []string{}
		actual := manifest[common.RAW]

		if a == (FailedTemplateValidator{}) && !context.Negative {
			// If the validator is empty and the context is not negative,
			// continue to the next iteration without throwing an error.
			continue
		}

		if a.ErrorPattern != "" {
			currentSuccess, validateSingleErrors = a.validateErrorPattern(actual, idx, context)
		} else if a.ErrorMessage != "" {
			currentSuccess, validateSingleErrors = a.validateErrorMessage(actual, idx, context)
		} else {
			currentSuccess = true
		}

		validateErrors = append(validateErrors, validateSingleErrors...)

		validateSuccess = determineSuccess(idx, validateSuccess, currentSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		validateSuccess = false
		errorMessage := a.failInfo("No failed document", -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}

func (a FailedTemplateValidator) validateErrorPattern(actual interface{}, idx int, context *ValidateContext) (bool, []string) {
	p, err := regexp.Compile(a.ErrorPattern)
	if err != nil {
		errorMessage := splitInfof(errorFormat, -1, err.Error())
		return false, errorMessage
	}

	if (actual != nil && p.MatchString(actual.(string))) == context.Negative {
		errorMessage := a.failInfo(actual, idx, context.Negative)
		return false, errorMessage
	}

	return true, []string{}
}

func (a FailedTemplateValidator) validateErrorMessage(actual interface{}, idx int, context *ValidateContext) (bool, []string) {
	if (actual != nil && reflect.DeepEqual(a.ErrorMessage, actual.(string))) == context.Negative {
		errorMessage := a.failInfo(actual, idx, context.Negative)
		return false, errorMessage
	}

	return true, []string{}
}
