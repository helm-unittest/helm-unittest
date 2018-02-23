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

func (a IsEmptyValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isEmptyFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be empty, got:
%s
`
	return splitInfof(isEmptyFailFormat, a.Path, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (a IsEmptyValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
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

	not := assert.IsNegative()
	if isEmpty != not {
		return true, []string{}
	}
	return false, a.failInfo(actual, not)
}
