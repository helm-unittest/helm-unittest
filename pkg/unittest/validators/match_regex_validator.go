package validators

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// MatchRegexValidator validate value of Path match Pattern
type MatchRegexValidator struct {
	Path         string
	Pattern      string
	DecodeBase64 bool
}

func (mrv MatchRegexValidator) failInfo(actual string, manifestIndex, actualIndex int, not bool) []string {

	log.WithField("validator", "match_regex").Debugln("expected pattern:", mrv.Pattern)
	log.WithField("validator", "match_regex").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, true, true, false, " to match"),
		manifestIndex,
		actualIndex,
		mrv.Path,
		mrv.Pattern,
		actual,
	)
}

func (mrv MatchRegexValidator) validateSingle(actual any, pattern *regexp.Regexp, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	if s, ok := actual.(string); ok {
		if mrv.DecodeBase64 {
			decodedSingleActual, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("unable to decode base64 expected content %s", actual))
			}
			s = string(decodedSingleActual)
		}

		if pattern.MatchString(s) == context.Negative {
			return false, mrv.failInfo(s, manifestIndex, actualIndex, context.Negative)
		} else {
			return true, []string{}
		}
	}

	return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf(
		"expect '%s' to be a string, got:\n%s",
		mrv.Path,
		common.TrustedMarshalYAML(actual),
	))
}

func (mrv MatchRegexValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, mrv.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	pattern, err := regexp.Compile(mrv.Pattern)
	if err != nil {
		return false, splitInfof(errorFormat, -1, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", mrv.Path))
	}

	validateManifestSuccess := (len(actuals) == 0 && context.Negative)
	var validateManifestErrors []string

	for actualIndex, actual := range actuals {
		validateSingleSuccess, validateSingleErrors := mrv.validateSingle(actual, pattern, manifestIndex, actualIndex, context)
		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)

		if !validateManifestSuccess && context.FailFast {
			break
		}
	}

	return validateManifestSuccess, validateManifestErrors
}

// Validate implement Validatable
func (mrv MatchRegexValidator) Validate(context *ValidateContext) (bool, []string) {
	verr := validateRequiredField(mrv.Pattern, "pattern")
	if verr != nil {
		return false, splitInfof(errorFormat, -1, -1, verr.Error())
	}

	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		currentSuccess, validateSingleErrors := mrv.validateManifest(manifest, manifestIndex, context)

		validateErrors = append(validateErrors, validateSingleErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, currentSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := mrv.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
