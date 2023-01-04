package validators

import (
	"strconv"

	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// MatchSnapshotValidator validate snapshot of value of Path the same as cached
type MatchSnapshotValidator struct {
	Path string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, index int, not bool) []string {
	customMessage := " to match snapshot " + strconv.Itoa(int(compared.Index))

	log.WithField("validator", "snapshot").Debugln("expected content:", compared.CachedSnapshot)
	log.WithField("validator", "snapshot").Debugln("actual content:", compared.NewSnapshot)

	var infoToShow string
	if not {
		infoToShow = compared.CachedSnapshot
	} else {
		infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
	}
	return splitInfof(
		setFailFormat(not, true, false, false, customMessage),
		index,
		v.Path,
		infoToShow,
	)
}

// Validate implement Validatable
func (v MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

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

		result := context.CompareToSnapshot(actual)

		if result.Passed == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(result, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
