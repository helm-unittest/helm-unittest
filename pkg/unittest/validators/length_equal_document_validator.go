package validators

import (
	"fmt"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
)

// LengthEqualDocumentsValidator validate whether the count of manifests rendered form template is Count
type LengthEqualDocumentsValidator struct {
	Paths []string // optional
	Path  string   // optional
	Count int      // optional if paths defined
}

func (v LengthEqualDocumentsValidator) singleValidateCounts(manifest common.K8sManifest, path string, idx, count int) (bool, []string, int) {
	spec, err := valueutils.GetValueOfSetPath(manifest, path)
	if err != nil {
		return false, splitInfof(errorFormat, idx, err.Error()), 0
	}
	specArr, ok := spec.([]interface{})
	if !ok {
		return false, splitInfof(errorFormat, idx, fmt.Sprintf("%s is not array", path)), 0
	}
	specLen := len(specArr)
	if count > -1 {
		if specLen != count {
			return false, splitInfof(errorFormat, idx, fmt.Sprintf(
				"count doesn't match. expected: %d != %d actual", count, specLen)), 0
		}
	}
	return true, []string{}, specLen
}

func (v LengthEqualDocumentsValidator) arraysValidateCounts(pathCount map[string]int, idx int) (bool, []string, int) {
	arrayCount := -1
	for k, pathCountValue := range pathCount {
		if arrayCount == -1 {
			arrayCount = pathCountValue
		} else if arrayCount != pathCountValue {
			arrayCount = -1
			return false, splitInfof(errorFormat, idx, fmt.Sprintf(
				"%s count is '%d'(doesn't match others)", k, pathCountValue)), arrayCount
		}
	}
	return true, []string{}, arrayCount
}

// Validate implement Validatable
func (v LengthEqualDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if len(v.Path) > 0 && v.Count == 0 {
		return false, splitInfof(errorFormat, -1, "'count' field must be set if 'path' is used")
	}
	if len(v.Path) > 0 && len(v.Paths) > 0 {
		return false, splitInfof(errorFormat, -1, "'paths' couldn't be used with 'path'")
	}
	singleMode := len(v.Path) > 0
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}
	validateSuccess := false
	validateErrors := make([]string, 0)
	for idx, manifest := range manifests {
		if singleMode {
			var validateSingleErrors []string
			validateSuccess, validateSingleErrors, _ = v.singleValidateCounts(manifest, v.Path, idx, v.Count)
			validateErrors = append(validateErrors, validateSingleErrors...)
			continue
		} else {
			pathCount := map[string]int{}
			optimizeCheck := true
			for _, path := range v.Paths {
				var validateSingleErrors []string
				validateSuccess, validateSingleErrors, pathCount[path] = v.singleValidateCounts(manifest, path, idx, -1)
				if !validateSuccess {
					validateErrors = append(validateErrors, validateSingleErrors...)
					optimizeCheck = false
				}
			}

			if !optimizeCheck {
				continue
			}

			var arrayCount int
			var validateSingleErrors []string
			validateSuccess, validateSingleErrors, arrayCount = v.arraysValidateCounts(pathCount, idx)
			validateErrors = append(validateErrors, validateSingleErrors...)

			if arrayCount == -1 {
				continue
			}
		}
		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}
	return validateSuccess, validateErrors
}
