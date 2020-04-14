package validators

import (
	"strconv"

	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// MatchSnapshotValidator validate snapshot of value of Path the same as cached
type MatchSnapshotValidator struct {
	Path string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, index int, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	snapshotFailFormat := `
Path:%s
Expected` + notAnnotation + ` to match snapshot ` + strconv.Itoa(int(compared.Index)) + `:
%s
`
	var infoToShow string
	if not {
		infoToShow = compared.CachedSnapshot
	} else {
		infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
	}
	return splitInfof(snapshotFailFormat, index, v.Path, infoToShow)
}

// Validate implement Validatable
func (v MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
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

		result := context.CompareToSnapshot(actual)

		if result.Passed == context.Negative {
			validateSuccess = validateSuccess && false
			errorMessage := v.failInfo(result, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = validateSuccess == true
	}

	return validateSuccess, validateErrors
}
