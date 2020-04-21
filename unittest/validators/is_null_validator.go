package validators

import (
	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// IsNullValidator validate value of Path id kind
type IsNullValidator struct {
	Path string
}

func (v IsNullValidator) failInfo(actual interface{}, index int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isNullFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be null, got:
%s
`
	return splitInfof(isNullFailFormat, index, v.Path, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (v IsNullValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if actual == nil == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(actual, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
