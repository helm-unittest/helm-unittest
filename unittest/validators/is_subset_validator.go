package validators

import (
	"fmt"
	"reflect"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"
)

// IsSubsetValidator validate whether value of Path contains Content
type IsSubsetValidator struct {
	Path    string
	Content interface{}
}

func (v IsSubsetValidator) failInfo(actual interface{}, index int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}
	containsFailFormat := `
Path:%s
Expected` + notAnnotation + ` to contain:
%s
Actual:
%s
`
	return splitInfof(
		containsFailFormat,
		index,
		v.Path,
		common.TrustedMarshalYAML(v.Content),
		common.TrustedMarshalYAML(actual),
	)
}

// Validate implement Validatable
func (v IsSubsetValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = validateSuccess && false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if actual, ok := actual.(map[interface{}]interface{}); ok {
			found := false

			for key, value := range actual {
				ele := map[interface{}]interface{}{key: value}
				if reflect.DeepEqual(ele, v.Content) {
					found = true
				}
			}

			if found == context.Negative {
				validateSuccess = validateSuccess && false
				errorMessage := v.failInfo(actual, idx, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}

			validateSuccess = validateSuccess && true
			continue
		}

		actualYAML, _ := yaml.Marshal(actual)
		validateSuccess = validateSuccess && false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect '%s' to be an object, got:\n%s",
			v.Path,
			string(actualYAML),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
