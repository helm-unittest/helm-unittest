package validators

import (
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// ExistsValidator validate value of Path id kind
type ExistsValidator struct {
	Path string
}

func (v ExistsValidator) failInfo(manifestIndex, actualIndex int, not bool) []string {
	format := "Path:%s expected to "

	if not {
		format = format + "NOT "
	}

	format = format + "exists"

	return splitInfof(
		format,
		manifestIndex,
		actualIndex,
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
			errorMessage := splitInfof(errorFormat, idx, -1, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		if len(actual) > 0 == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(idx, -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo(-1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
