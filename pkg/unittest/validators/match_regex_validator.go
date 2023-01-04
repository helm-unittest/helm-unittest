package validators

import (
	"fmt"
	"regexp"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// MatchRegexValidator validate value of Path match Pattern
type MatchRegexValidator struct {
	Path    string
	Pattern string
}

func (v MatchRegexValidator) failInfo(actual string, index int, not bool) []string {

	log.WithField("validator", "match_regex").Debugln("expected pattern:", v.Pattern)
	log.WithField("validator", "match_regex").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, true, true, false, " to match"),
		index,
		v.Path,
		v.Pattern,
		actual,
	)
}

// Validate implement Validatable
func (v MatchRegexValidator) Validate(context *ValidateContext) (bool, []string) {
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
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		p, err := regexp.Compile(v.Pattern)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, -1, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			break
		}

		if s, ok := actual.(string); ok {
			if p.MatchString(s) == context.Negative {
				validateSuccess = false
				errorMessage := v.failInfo(s, idx, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}

			validateSuccess = determineSuccess(idx, validateSuccess, true)
			continue
		}

		validateSuccess = false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect '%s' to be a string, got:\n%s",
			v.Path,
			common.TrustedMarshalYAML(actual),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
