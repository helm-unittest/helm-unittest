package validators

import (
	"strconv"

	"fmt"
	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// MatchSnapshotValidator validate snapshot of value of Path the same as cached
type MatchSnapshotValidator struct {
	Path string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, not bool) []string {
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
	return splitInfof(snapshotFailFormat, v.Path, infoToShow)
}

// Validate implement Validatable
func (v MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest, err := context.getManifest()
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	allManifests, err := context.getAllManifests()
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}
	var allActual []common.K8sManifest

	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if v.Path == "ALL" {
		allActual = allManifests
	}
	// fmt.Println("Start")
	fmt.Println(allActual)
	// fmt.Println("End")

	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	var result *snapshot.CompareResult

	if v.Path == "ALL" {
		result = context.CompareToSnapshot(allActual)
	}else {
		result = context.CompareToSnapshot(actual);	
	}

	if result.Passed != context.Negative {
		return true, []string{}
	}
	return false, v.failInfo(result, context.Negative)
}
