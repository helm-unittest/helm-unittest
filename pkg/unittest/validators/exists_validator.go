package validators

import (
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// ExistsValidator validate value of Path id kind
type ExistsValidator struct {
	Path string
}

func (v ExistsValidator) failInfo(index int, not bool) []string {
	format := "Path:%s expected to "

	if not {
		format = format + "NOT "
	}

	format = format + "exists"

	return splitInfof(
		format,
		index,
		v.Path,
	)
}

// Validate implement Validatable
func (v ExistsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		if len(actual) > 0 == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
