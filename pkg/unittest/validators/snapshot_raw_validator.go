package validators

import (
	"strconv"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	log "github.com/sirupsen/logrus"
)

// MatchSnapshotRawValidator validate snapshot of value of Path the same as cached
type MatchSnapshotRawValidator struct{}

func (v MatchSnapshotRawValidator) failInfo(compared *snapshot.CompareResult, not bool) []string {
	customMessage := " to match snapshot " + strconv.Itoa(int(compared.Index))

	log.WithField("validator", "snapshot_raw").Debugln("expected content:", compared.CachedSnapshot)
	log.WithField("validator", "snapshot_raw").Debugln("actual content:", compared.NewSnapshot)

	var infoToShow string
	if not {
		infoToShow = compared.CachedSnapshot
	} else {
		infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
	}
	return splitInfof(
		setFailFormat(not, false, false, false, customMessage),
		-1,
		infoToShow,
	)
}

// Validate implement Validatable
func (v MatchSnapshotRawValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		result := context.CompareToSnapshot(actual)

		if result.Passed == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(result, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
