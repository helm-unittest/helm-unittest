package validators

import (
	"regexp"

	"github.com/helm-unittest/helm-unittest/internal/common"
	log "github.com/sirupsen/logrus"
)

// MatchRegexRawValidator validate value of Path match Pattern
type MatchRegexRawValidator struct {
	Pattern string
}

func (mrrv MatchRegexRawValidator) failInfo(actual string, not bool) []string {

	log.WithField("validator", "match_regex_raw").Debugln("expected pattern:", mrrv.Pattern)
	log.WithField("validator", "match_regex_raw").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, false, true, false, " to match"),
		-1,
		-1,
		mrrv.Pattern,
		actual,
	)
}

// Validate implement Validatable
func (mrrv MatchRegexRawValidator) Validate(context *ValidateContext) (bool, []string) {
	verr := validateRequiredField(mrrv.Pattern, "pattern")
	if verr != nil {
		return false, splitInfof(errorFormat, -1, -1, verr.Error())
	}

	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		p, err := regexp.Compile(mrrv.Pattern)
		if err != nil {
			return false, splitInfof(errorFormat, -1, -1, err.Error())
		}

		if p.MatchString(actual) == context.Negative {
			validateSuccess = false
			errorMessage := mrrv.failInfo(actual, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)

			if context.FailFast {
				break
			}
			continue
		}

		validateSuccess = determineSuccess(manifestIndex, validateSuccess, true)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := mrrv.failInfo("no manifest found", context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
