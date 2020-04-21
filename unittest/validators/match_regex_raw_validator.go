package validators

import (
	"regexp"

	"github.com/lrills/helm-unittest/unittest/common"
)

// MatchRegexRawValidator validate value of Path match Pattern
type MatchRegexRawValidator struct {
	Pattern string
}

func (v MatchRegexRawValidator) failInfo(actual string, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	regexFailFormat := `
Expected` + notAnnotation + ` to match:%s
Actual:%s
`
	return splitInfof(regexFailFormat, -1, v.Pattern, actual)
}

// Validate implement Validatable
func (v MatchRegexRawValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
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

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
