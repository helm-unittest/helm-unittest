package validators

import (
	"fmt"
	"reflect"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// IsEmptyValidator validate value of Path is empty
type IsEmptyValidator struct {
	Path string
}

func (v IsEmptyValidator) failInfo(actual interface{}, index int, not bool) []string {
	actualYAML := common.TrustedMarshalYAML(actual)

	log.WithField("validator", "is_empty").Debugln("actual content:", actualYAML)

	return splitInfof(
		setFailFormat(not, true, false, false, " to be empty, got"),
		index,
		v.Path,
		actualYAML,
	)
}

// Validate implement Validatable
func (v IsEmptyValidator) Validate(context *ValidateContext) (bool, []string) {
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

		if len(actual) == 0 {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown path %s", v.Path))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		singleValue := actual[0]
		actualValue := reflect.ValueOf(singleValue)
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
			errorMessage := v.failInfo(singleValue, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
