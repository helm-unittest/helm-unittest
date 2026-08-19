package validators

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// ExistsValidator validate value of Path id kind
type ExistsValidator struct {
	Path         string
	ParseOptions `mapstructure:",squash"`
}

func (v ExistsValidator) failInfo(manifestIndex, actualIndex int, not bool) []string {
	format := "Path:%s expected to "

	if not {
		format = format + "NOT "
	}

	format = format + "exists"

	return splitInfof(
		format,
		manifestIndex,
		actualIndex,
		v.Path,
	)
}

// validateManifest reports whether the path resolved to at least one value.
//
// When parsing is requested, existence is decided by whether the innerPath
// resolved within the parsed content, not by whether the outer path resolved.
// An empty result is therefore a legitimate "does not exist" answer rather than
// an error, which is what makes `exists` with `innerPath` meaningful.
//
// The third return value distinguishes a path/parse error from a failed
// existence check, so Validate can preserve the original control flow: only
// an error is eligible to break the loop under FailFast; a failed existence
// check never is.
func (v ExistsValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string, bool) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error()), true
	}

	actual, err = v.ParseOptions.resolveActuals(actual, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error()), true
	}

	if len(actual) > 0 == context.Negative {
		return false, v.failInfo(manifestIndex, -1, context.Negative), false
	}

	return true, []string{}, false
}

// Validate implement Validatable
func (v ExistsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		manifestSuccess, manifestValidateErrors, isError := v.validateManifest(manifest, idx, context)
		if !manifestSuccess {
			validateSuccess = false
			validateErrors = append(validateErrors, manifestValidateErrors...)
			if isError && context.FailFast {
				break
			}
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo(-1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
