package validators

import (
	"fmt"
	"reflect"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"
)

// ContainsValidator validate whether value of Path is an array and contains Content
type ContainsValidator struct {
	Path    string
	Content interface{}
	Count   *int
	Any     bool
}

func (v ContainsValidator) failInfo(actual interface{}, index int, not bool) []string {
	containsFailFormat := setFailFormat(not, true, true, false, " to contain")

	return splitInfof(
		containsFailFormat,
		index,
		v.Path,
		common.TrustedMarshalYAML([]interface{}{v.Content}),
		common.TrustedMarshalYAML(actual),
	)
}

func (v ContainsValidator) validateContent(actual []interface{}) (bool, int) {
	found := false
	validateFoundCount := 0

	for _, ele := range actual {
		// When any enabled, only the key is validated
		if v.Any {
			if subset, ok := ele.(map[interface{}]interface{}); ok {
				if validateSubset(subset, v.Content) {
					found = true
					validateFoundCount++
				}
			}
		}

		if !v.Any && reflect.DeepEqual(ele, v.Content) {
			found = true
			validateFoundCount++
		}
	}

	return found, validateFoundCount
}

// Validate implement Validatable
func (v ContainsValidator) Validate(context *ValidateContext) (bool, []string) {
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

		if actual, ok := actual.([]interface{}); ok {
			found, validateFoundCount := v.validateContent(actual)

			// no found, regardless count, inverse awareness
			if found == context.Negative {
				if v.Count == nil ||
					(v.Count != nil && *v.Count == validateFoundCount && context.Negative) ||
					(v.Count != nil && *v.Count != validateFoundCount && !context.Negative) {
					validateSuccess = false
					errorMessage := v.failInfo(actual, idx, context.Negative)
					validateErrors = append(validateErrors, errorMessage...)
					continue
				}
			}

			// invalid count, found
			// valid count (so found), invalid found
			if v.Count != nil && found != context.Negative {
				if (*v.Count != validateFoundCount && !context.Negative) ||
					(*v.Count == validateFoundCount && context.Negative) {
					actualYAML, _ := yaml.Marshal(actual)
					validateSuccess = false
					errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
						"expect count %d in '%s' to be in array, got %d:\n%s",
						*v.Count,
						v.Path,
						validateFoundCount,
						string(actualYAML),
					))
					validateErrors = append(validateErrors, errorMessage...)
					continue
				}
			}

			validateSuccess = determineSuccess(idx, validateSuccess, true)
			continue
		}

		actualYAML, _ := yaml.Marshal(actual)
		validateSuccess = false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect '%s' to be an array, got:\n%s",
			v.Path,
			string(actualYAML),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
