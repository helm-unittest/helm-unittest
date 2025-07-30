package validators

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// LengthEqualDocumentsValidator validate whether the count of manifests rendered form template is Count
type LengthEqualDocumentsValidator struct {
	Paths []string // optional
	Path  string   // optional
	Count *int     // optional if paths defined
}

func (v LengthEqualDocumentsValidator) failInfo(path, count, actual string, manifestIndex, valueIndex int, not bool) []string {
	customMessage := " to match count"

	return splitInfof(
		setFailFormat(not, true, true, false, customMessage),
		manifestIndex,
		valueIndex,
		path,
		count,
		actual)
}

func (v LengthEqualDocumentsValidator) singleValidateCounts(manifest common.K8sManifest, path string, manifestIndex int, context *ValidateContext) (bool, []string, int) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error()), 0
	}

	manifestSuccess := (len(actuals) == 0 && context.Negative)
	var manifestErrors []string
	arrayLength := 0

	for actualIndex, actual := range actuals {
		actualSuccess := false
		actualArray, ok := actual.([]any)
		if !ok && v.Count == nil {
			actualErrors := splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("%s is not array", path))
			manifestErrors = append(manifestErrors, actualErrors...)
			continue
		}
		arrayLength = len(actualArray)
		if v.Count != nil && (arrayLength == *v.Count == context.Negative) {
			actualErrors := v.failInfo(path, strconv.Itoa(*v.Count), strconv.Itoa(arrayLength), manifestIndex, -1, context.Negative)
			manifestErrors = append(manifestErrors, actualErrors...)
		} else {
			actualSuccess = true
		}

		manifestSuccess = determineSuccess(actualIndex, manifestSuccess, actualSuccess)

		if !manifestSuccess && context.FailFast {
			break
		}
	}

	return manifestSuccess, manifestErrors, arrayLength
}

func (v LengthEqualDocumentsValidator) arraysValidateCounts(pathCount map[string]int, manifestIndex int, context *ValidateContext) (bool, []string, int) {
	arrayCount := -1

	// Sort alphabetically to get a standardized result
	pathSlice := make([]string, 0)
	for path := range pathCount {
		pathSlice = append(pathSlice, path)
	}

	sort.Strings(pathSlice)

	for _, path := range pathSlice {
		pathCountValue := pathCount[path]
		if arrayCount == -1 {
			arrayCount = pathCountValue
		} else if (arrayCount == pathCountValue) == context.Negative {
			arrayCount = -1
			return false, v.failInfo(path, "-1", strconv.Itoa(pathCountValue), manifestIndex, -1, context.Negative), arrayCount
		}
	}

	return true, []string{}, arrayCount
}

func (v LengthEqualDocumentsValidator) validatePathCount() bool {
	return len(v.Path) > 0 && (v.Count == nil || (v.Count != nil && *v.Count < 0))
}

func (v LengthEqualDocumentsValidator) validatePathPaths() bool {
	return len(v.Path) > 0 && len(v.Paths) > 0
}

func (v LengthEqualDocumentsValidator) validateSingleMode(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	validateSuccess, validateSingleErrors, _ := v.singleValidateCounts(manifest, v.Path, manifestIndex, context)
	return validateSuccess, validateSingleErrors
}

func (v LengthEqualDocumentsValidator) validateMultipleMode(manifest common.K8sManifest, idx int, context *ValidateContext) (bool, []string) {
	var manifestErrors []string
	var validateSingleErrors []string
	pathCount := map[string]int{}
	optimizeCheck := true
	validateSuccess := false

	for _, path := range v.Paths {
		validateSuccess, validateSingleErrors, pathCount[path] = v.singleValidateCounts(manifest, path, idx, context)
		if !validateSuccess {
			manifestErrors = append(manifestErrors, validateSingleErrors...)
			optimizeCheck = false
		}
	}

	if !optimizeCheck {
		return false, manifestErrors
	}

	var arrayCount int
	validateSuccess, validateSingleErrors, arrayCount = v.arraysValidateCounts(pathCount, idx, context)
	if !validateSuccess {
		manifestErrors = append(manifestErrors, validateSingleErrors...)
	}

	if arrayCount == -1 {
		return false, manifestErrors
	}

	return validateSuccess, manifestErrors
}

// Validate implement Validatable
func (v LengthEqualDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if v.validatePathCount() {
		return false, splitInfof(errorFormat, -1, -1, "'count' field must be set if 'path' is used")
	}
	if v.validatePathPaths() {
		return false, splitInfof(errorFormat, -1, -1, "'paths' couldn't be used with 'path'")
	}

	singleMode := len(v.Path) > 0
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		currentSuccess := false
		var validateManifestErrors []string
		if singleMode {
			currentSuccess, validateManifestErrors = v.validateSingleMode(manifest, manifestIndex, context)
		} else {
			currentSuccess, validateManifestErrors = v.validateMultipleMode(manifest, manifestIndex, context)
		}

		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, currentSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo("", "", "no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
