package validators

import (
	"regexp"

	"github.com/lrills/helm-unittest/internal/common"
	log "github.com/sirupsen/logrus"
)

// MatchRegexRawValidator validate value of Path match Pattern
type MatchRegexRawValidator struct {
	Pattern string
}

func (v MatchRegexRawValidator) failInfo(actual string, not bool) []string {

	log.WithField("validator", "match_regex_raw").Debugln("expected pattern:", v.Pattern)
	log.WithField("validator", "match_regex_raw").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, false, true, false, " to match"),
		-1,
		v.Pattern,
		actual,
	)
}

// Validate implement Validatable
func (v MatchRegexRawValidator) Validate(context *ValidateContext) (bool, []string) {
	verr := validateRequiredField(v.Pattern, "pattern")
	if verr != nil {
		return false, splitInfof(errorFormat, -1, verr.Error())
	}

	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		p, err := regexp.Compile(v.Pattern)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, -1, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			break
		}

		if p.MatchString(actual) == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(actual, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
