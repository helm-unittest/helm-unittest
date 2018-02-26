package validators

import (
	"strconv"

	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// MatchSnapshotValidator
type MatchSnapshotValidator struct {
	Path string
}

func (a MatchSnapshotValidator) failInfo(compared *snapshot.SnapshotCompareResult, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	snapshotFailFormat := `
Path:%s
Expected` + notAnnotation + ` to match snapshot ` + strconv.Itoa(compared.Index) + `:
%s
`
	var infoToShow string
	if not {
		infoToShow = compared.Cached
	} else {
		infoToShow = diff(compared.Cached, compared.New)
	}
	return splitInfof(snapshotFailFormat, a.Path, infoToShow)
}

// Validate implement Validatable
func (a MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest := context.Docs[context.Index]

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	result := context.CompareToSnapshot(actual)

	if result.Matched != context.Negative {
		return true, []string{}
	}
	return false, a.failInfo(result, context.Negative)
}
