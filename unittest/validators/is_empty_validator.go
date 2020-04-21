package validators

import (
	"reflect"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// IsEmptyValidator validate value of Path is empty
type IsEmptyValidator struct {
	Path string
}

func (v IsEmptyValidator) failInfo(actual interface{}, index int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isEmptyFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be empty, got:
%s
`
	return splitInfof(isEmptyFailFormat, index, v.Path, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (v IsEmptyValidator) Validate(context *ValidateContext) (bool, []string) {
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

		actualValue := reflect.ValueOf(actual)
		var isEmpty bool
		switch actualValue.Kind() {
		case reflect.Invalid:
			isEmpty = true
		case reflect.Array, reflect.Map, reflect.Slice:
			isEmpty = actualValue.Len() == 0
		default:
			zero := reflect.Zero(actualValue.Type())
			isEmpty = reflect.DeepEqual(actual, zero.Interface())
		}

		if isEmpty == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(actual, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
