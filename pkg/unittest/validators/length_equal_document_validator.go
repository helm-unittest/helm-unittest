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
	Paths        []string // optional
	Path         string   // optional
	Count        *int     // optional if paths defined
	ParseOptions `mapstructure:",squash"`
}

func (ledv LengthEqualDocumentsValidator) failInfo(path, count, actual string, manifestIndex, valueIndex int, not bool) []string {
	customMessage := " to match count"

	return splitInfof(
		setFailFormat(not, true, true, false, customMessage),
		manifestIndex,
		valueIndex,
		path,
		count,
		actual)
}

func (ledv LengthEqualDocumentsValidator) singleValidateCounts(manifest common.K8sManifest, path string, manifestIndex int, context *ValidateContext) (bool, []string, int) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error()), 0
	}

	actuals, err = ledv.resolveActuals(actuals, path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error()), 0
	}

	manifestSuccess := (len(actuals) == 0 && context.Negative)
	var manifestErrors []string
	arrayLength := 0

	for actualIndex, actual := range actuals {
		actualSuccess := false
		actualArray, ok := actual.([]any)
		if !ok && ledv.Count == nil {
			actualErrors := splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("%s is not array", path))
			manifestErrors = append(manifestErrors, actualErrors...)
			continue
		}
		arrayLength = len(actualArray)
		if ledv.Count != nil && (arrayLength == *ledv.Count == context.Negative) {
			actualErrors := ledv.failInfo(path, strconv.Itoa(*ledv.Count), strconv.Itoa(arrayLength), manifestIndex, -1, context.Negative)
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

func (ledv LengthEqualDocumentsValidator) arraysValidateCounts(pathCount map[string]int, manifestIndex int, context *ValidateContext) (bool, []string, int) {
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
			return false, ledv.failInfo(path, "-1", strconv.Itoa(pathCountValue), manifestIndex, -1, context.Negative), arrayCount
		}
	}

	return true, []string{}, arrayCount
}

func (ledv LengthEqualDocumentsValidator) validatePathCount() bool {
	return len(ledv.Path) > 0 && (ledv.Count == nil || (ledv.Count != nil && *ledv.Count < 0))
}

func (ledv LengthEqualDocumentsValidator) validatePathPaths() bool {
	return len(ledv.Path) > 0 && len(ledv.Paths) > 0
}

func (ledv LengthEqualDocumentsValidator) validateSingleMode(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	validateSuccess, validateSingleErrors, _ := ledv.singleValidateCounts(manifest, ledv.Path, manifestIndex, context)
	return validateSuccess, validateSingleErrors
}

func (ledv LengthEqualDocumentsValidator) validateMultipleMode(manifest common.K8sManifest, idx int, context *ValidateContext) (bool, []string) {
	var manifestErrors []string
	var validateSingleErrors []string
	pathCount := map[string]int{}
	optimizeCheck := true
	validateSuccess := false

	for _, path := range ledv.Paths {
		validateSuccess, validateSingleErrors, pathCount[path] = ledv.singleValidateCounts(manifest, path, idx, context)
		if !validateSuccess {
			manifestErrors = append(manifestErrors, validateSingleErrors...)
			optimizeCheck = false
		}
	}

	if !optimizeCheck {
		return false, manifestErrors
	}

	var arrayCount int
	validateSuccess, validateSingleErrors, arrayCount = ledv.arraysValidateCounts(pathCount, idx, context)
	if !validateSuccess {
		manifestErrors = append(manifestErrors, validateSingleErrors...)
	}

	if arrayCount == -1 {
		return false, manifestErrors
	}

	return validateSuccess, manifestErrors
}

// Validate implement Validatable
func (ledv LengthEqualDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if ledv.validatePathCount() {
		return false, splitInfof(errorFormat, -1, -1, "'count' field must be set if 'path' is used")
	}
	if ledv.validatePathPaths() {
		return false, splitInfof(errorFormat, -1, -1, "'paths' couldn't be used with 'path'")
	}

	singleMode := len(ledv.Path) > 0
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		currentSuccess := false
		var validateManifestErrors []string
		if singleMode {
			currentSuccess, validateManifestErrors = ledv.validateSingleMode(manifest, manifestIndex, context)
		} else {
			currentSuccess, validateManifestErrors = ledv.validateMultipleMode(manifest, manifestIndex, context)
		}

		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, currentSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := ledv.failInfo("", "", "no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
