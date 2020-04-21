package validators

import "github.com/lrills/helm-unittest/unittest/common"

// IsKindValidator validate kind of manifest is Of
type IsKindValidator struct {
	Of string
}

func (v IsKindValidator) failInfo(actual interface{}, index int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	isKindFailFormat := "Expected" + notAnnotation + " kind:%s"
	if not {
		return splitInfof(isKindFailFormat, index, v.Of)
	}
	return splitInfof(isKindFailFormat+"\nActual:%s", index, v.Of, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (v IsKindValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		if kind, ok := manifest["kind"].(string); (ok && kind == v.Of) == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(manifest["kind"], idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
