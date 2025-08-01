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
	Content any
	Count   *int
	Any     bool
}

func (v ContainsValidator) failInfo(actual any, manifestIndex, assertIndex int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML([]any{v.Content})
	actualYAML := common.TrustedMarshalYAML(actual)
	containsFailFormat := setFailFormat(not, true, true, false, " to contain")

	log.WithField("validator", "contains").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "contains").Debugln("actual content:", actualYAML)

	return splitInfof(
		containsFailFormat,
		manifestIndex,
		assertIndex,
		v.Path,
		expectedYAML,
		actualYAML,
	)
}

func (v ContainsValidator) validateContent(actual []any) (bool, int) {
	found := false
	validateFoundCount := 0

	for _, ele := range actual {
		isArray, isSubmatch := v.isArrayOrSubsetMatch(ele)
		if isSubmatch {
			found = true
			validateFoundCount++
		}

		if v.isExactMatch(isArray, ele) {
			found = true
			validateFoundCount++
		}
	}

	return found, validateFoundCount
}

func (v ContainsValidator) isArrayOrSubsetMatch(ele any) (bool, bool) {
	if v.Any {
		subset, subsetOk := ele.(map[string]any)
		content, contentOk := v.Content.(map[string]any)
		if subsetOk && contentOk {
			return false, validateSubset(subset, content)
		} else {
			return true, false
		}
	}

	return false, false
}

func (v ContainsValidator) isExactMatch(isArray bool, ele any) bool {
	return (!v.Any || isArray) && reflect.DeepEqual(ele, v.Content)
}

func (v ContainsValidator) validateFoundCount(validateFoundCount int) bool {
	return v.Count != nil && *v.Count == validateFoundCount
}

func (v ContainsValidator) validateSingle(singleActual []any, manifestIndex, assertIndex int, context *ValidateContext) (bool, []string) {
	validateSingleErrors := []string{}
	found, validateFoundCount := v.validateContent(singleActual)

	if v.Count == nil && (found == context.Negative) {
		validateSingleErrors = v.failInfo(singleActual, manifestIndex, assertIndex, context.Negative)
		return false, validateSingleErrors
	}

	// Found so check if the count is correct
	if v.Count != nil && ((found && v.validateFoundCount(validateFoundCount)) == context.Negative) {
		actualYAML := common.TrustedMarshalYAML(singleActual)
		if !found {
			validateSingleErrors = v.failInfo(singleActual, manifestIndex, assertIndex, context.Negative)
		} else {
			validateSingleErrors = splitInfof(errorFormat, manifestIndex, assertIndex, fmt.Sprintf(
				"expect count %d in '%s' to be in array, got %d:\n%s",
				*v.Count,
				v.Path,
				validateFoundCount,
				actualYAML,
			))
		}
		return false, validateSingleErrors
	}

	return true, validateSingleErrors
}

func (v ContainsValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	manifestSuccess := (len(actual) == 0 && context.Negative)
	var manifestValidateErrors []string

	for valuesIndex, singleActual := range actual {
		singleSuccess := false
		var singleValidateErrors []string
		convertedSingleActual, ok := singleActual.([]any)
		if ok {
			singleSuccess, singleValidateErrors = v.validateSingle(convertedSingleActual, manifestIndex, valuesIndex, context)
		} else {
			actualYAML := common.TrustedMarshalYAML(singleActual)
			singleValidateErrors = splitInfof(errorFormat, manifestIndex, valuesIndex, fmt.Sprintf(
				"expect '%s' to be an array, got:\n%s",
				v.Path,
				actualYAML,
			))
		}

		manifestValidateErrors = append(manifestValidateErrors, singleValidateErrors...)
		manifestSuccess = determineSuccess(valuesIndex, manifestSuccess, singleSuccess)

		if !manifestSuccess && context.FailFast {
			break
		}
	}

	return manifestSuccess, manifestValidateErrors
}

// Validate implement Validatable
func (v ContainsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		manifestSuccess, manifestValidateErrors := v.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, manifestValidateErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, manifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo("no manifest found", 0, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
