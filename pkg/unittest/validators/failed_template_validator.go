package validators

import (
	"cmp"
	"fmt"
	"reflect"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

const metaCharactersPattern = `.*[\\.+*?()|[\]{}^$].*`

var metaRegex = regexp.MustCompile(metaCharactersPattern)

// FailedTemplateValidator validate whether the errorMessage equal to errorMessage
type FailedTemplateValidator struct {
	ErrorMessage string
	ErrorPattern string
}

func (ftv FailedTemplateValidator) failInfo(actual any, manifestIndex, actualIndex int, not bool) []string {
	customMessage := " to equal"
	if ftv.ErrorPattern != "" {
		customMessage = " to match"
	}

	message := cmp.Or(ftv.ErrorMessage, ftv.ErrorPattern)

	log.WithField("validator", "failed_template").Debugln("expected content:", message)
	log.WithField("validator", "failed_template").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			manifestIndex,
			actualIndex,
			message,
		)
	}

	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		manifestIndex,
		actualIndex,
		message,
		fmt.Sprintf("%s", actual),
	)
}

func (ftv FailedTemplateValidator) validateManifests(manifests []common.K8sManifest, context *ValidateContext) (bool, []string) {
	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		currentSuccess := false
		var validateSingleErrors []string
		actual := manifest[common.RAW]

		if ftv == (FailedTemplateValidator{}) && !context.Negative {
			// If the validator is empty and the context is not negative,
			// continue to the next iteration without throwing an error.
			continue
		}

		if ftv.ErrorPattern != "" {
			currentSuccess, validateSingleErrors = ftv.validateErrorPattern(actual, idx, -1, context)
		} else if ftv.ErrorMessage != "" {
			currentSuccess, validateSingleErrors = ftv.validateErrorMessage(actual, idx, -1, context)
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
		errorMessage := ftv.failInfo("No failed document", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}

func (ftv FailedTemplateValidator) validateErrorPattern(actual any, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	p, err := compilePattern(ftv.ErrorPattern)
	if err != nil {
		return false, splitInfof(errorFormat, -1, -1, err.Error())
	}
	if actual != nil && p.MatchString(actual.(string)) == context.Negative {
		if metaRegex.MatchString(ftv.ErrorPattern) {
			return handleMetaCharacters(ftv, actual, manifestIndex, actualIndex, context)
		}
		return false, ftv.failInfo(actual, manifestIndex, actualIndex, context.Negative)
	}
	return true, []string{}
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	p, err := regexp.Compile(pattern)
	if err != nil {
		escaped := regexp.QuoteMeta(pattern)
		log.WithField("validator", "failed_template").Debugln("fallback to escaped regex", escaped)
		p, err = regexp.Compile(escaped)
	}
	return p, err
}

func handleMetaCharacters(ftv FailedTemplateValidator, actual any, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	escaped := regexp.QuoteMeta(ftv.ErrorPattern)
	log.WithField("validator", "failed_template").Debugln("fallback to escaped regex", escaped)
	p, err := regexp.Compile(escaped)
	if err != nil {
		log.WithField("validator", "failed_template").Debugln(fmt.Sprintf("failed to regexp.Compile(%s)", escaped))
		return false, splitInfof(errorFormat, -1, -1, err.Error())
	}
	if p.MatchString(actual.(string)) == context.Negative {
		return false, ftv.failInfo(actual, manifestIndex, actualIndex, context.Negative)
	}
	return true, []string{}
}

func (ftv FailedTemplateValidator) validateErrorMessage(actual any, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	if (actual != nil && reflect.DeepEqual(ftv.ErrorMessage, actual.(string))) == context.Negative {
		errorMessage := ftv.failInfo(actual, manifestIndex, actualIndex, context.Negative)
		return false, errorMessage
	}

	return true, []string{}
}

// Validate implement Validatable
func (ftv FailedTemplateValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()
	validateSuccess := false
	validateErrors := make([]string, 0)

	if ftv.ErrorMessage != "" && ftv.ErrorPattern != "" {
		errorMessage := splitInfof(errorFormat, -1, -1, "single attribute 'errorMessage' or 'errorPattern' supported at the same time")
		validateErrors = append(validateErrors, errorMessage...)
		return false, validateErrors
	}

	if context.RenderError != nil {
		// Validating error, when the errorSource is due to rendering errors
		if ftv.ErrorPattern != "" {
			return ftv.validateErrorPattern(context.RenderError.Error(), -1, -1, context)
		} else if ftv.ErrorMessage != "" {
			return ftv.validateErrorMessage(context.RenderError.Error(), -1, -1, context)
		} else {
			validateSuccess = true
		}
	} else {
		var errorsToAppend []string
		validateSuccess, errorsToAppend = ftv.validateManifests(manifests, context)
		validateErrors = append(validateErrors, errorsToAppend...)
	}

	return validateSuccess, validateErrors
}
