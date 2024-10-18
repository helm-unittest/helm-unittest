package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// ContainsValidator validate whether value of Path is an array and contains Content
type ContainsValidator struct {
	Path    string
	Content interface{}
	Count   *int
	Any     bool
}

func (v ContainsValidator) failInfo(actual interface{}, manifestIndex int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML([]interface{}{v.Content})
	actualYAML := common.TrustedMarshalYAML(actual)
	containsFailFormat := setFailFormat(not, true, true, false, " to contain")

	log.WithField("validator", "contains").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "contains").Debugln("actual content:", actualYAML)

	return splitInfof(
		containsFailFormat,
		manifestIndex,
		v.Path,
		expectedYAML,
		actualYAML,
	)
}

func (v ContainsValidator) validateContent(actual []interface{}) (bool, int) {
	found := false
	validateFoundCount := 0

	for _, ele := range actual {
		// When any enabled, only the key is validated
		if v.Any {
			subset, subsetOk := ele.(map[string]interface{})
			content, contentOk := v.Content.(map[string]interface{})
			if subsetOk && contentOk {
				if validateSubset(subset, content) {
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

func (v ContainsValidator) validateFound(found, negative bool, validateFoundCount int) bool {
	return found == negative && (v.Count == nil ||
		(v.Count != nil && *v.Count == validateFoundCount && negative) ||
		(v.Count != nil && *v.Count != validateFoundCount && !negative))
}

func (v ContainsValidator) validateFoundCount(found, negative bool, validateFoundCount int) bool {
	return (v.Count != nil && found != negative) &&
		((*v.Count != validateFoundCount && !negative) ||
			(*v.Count == validateFoundCount && negative))
}

func (v ContainsValidator) validateSingle(singleActual []interface{}, idx int, context *ValidateContext) (bool, []string) {
	found, validateFoundCount := v.validateContent(singleActual)

	// no found, regardless count, inverse awareness
	if v.validateFound(found, context.Negative, validateFoundCount) {
		validateSingleErrors := v.failInfo(singleActual, idx, context.Negative)
		return false, validateSingleErrors
	}

	// invalid count, found
	// valid count (so found), invalid found
	if v.validateFoundCount(found, context.Negative, validateFoundCount) {
		actualYAML := common.TrustedMarshalYAML(singleActual)
		validateSingleErrors := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect count %d in '%s' to be in array, got %d:\n%s",
			*v.Count,
			v.Path,
			validateFoundCount,
			actualYAML,
		))
		return false, validateSingleErrors
	}

	return true, []string{}
}

// Validate implement Validatable
func (v ContainsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

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

		var manifestValidateErrors []string
		manifestSuccess := false

		for valuesIndex, singleActual := range actual {
			convertedSingleActual, ok := singleActual.([]interface{})
			if ok {
				manifestSuccess, manifestValidateErrors = v.validateSingle(convertedSingleActual, idx, context)
			} else {
				actualYAML := common.TrustedMarshalYAML(singleActual)
				manifestValidateErrors = splitInfof(errorFormat, idx, fmt.Sprintf(
					"expect '%s' to be an array, got:\n%s",
					v.Path,
					actualYAML,
				))
			}

			validateErrors = append(validateErrors, manifestValidateErrors...)
			validateSuccess = determineSuccess(valuesIndex, validateSuccess, manifestSuccess)
		}
	}

	return validateSuccess, validateErrors
}
