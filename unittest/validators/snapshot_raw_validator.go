package validators

import (
	"strconv"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
)

// MatchSnapshotRawValidator validate snapshot of value of Path the same as cached
type MatchSnapshotRawValidator struct{}

func (v MatchSnapshotRawValidator) failInfo(compared *snapshot.CompareResult, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	snapshotFailFormat := `
Expected` + notAnnotation + ` to match snapshot ` + strconv.Itoa(int(compared.Index)) + `:
%s
`
	var infoToShow string
	if not {
		infoToShow = compared.CachedSnapshot
	} else {
		infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
	}
	return splitInfof(snapshotFailFormat, -1, infoToShow)
}

// Validate implement Validatable
func (v MatchSnapshotRawValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		result := context.CompareToSnapshot(actual)

		if result.Passed == context.Negative {
			validateSuccess = validateSuccess && false
			errorMessage := v.failInfo(result, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = validateSuccess == true
	}

	return validateSuccess, validateErrors
}
