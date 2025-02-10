package validators

import (
	"fmt"
	"strconv"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// MatchSnapshotValidator validate snapshot of value of Path the same as cached
type MatchSnapshotValidator struct {
	Path string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, manifestIndex, actualIndex int, not bool) []string {
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
		manifestIndex,
		actualIndex,
		v.Path,
		infoToShow,
	)
}

func (v MatchSnapshotValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	validateManifestSuccess := (len(actual) == 0 && context.Negative)
	var validateManifestErrors []string

	for actualIndex, singleActual := range actual {
		validateSingleSuccess := false
		var validateSingleErrors []string
		result := context.CompareToSnapshot(singleActual)

		if result.Passed == context.Negative {
			validateSingleErrors = v.failInfo(result, manifestIndex, actualIndex, context.Negative)
		} else {
			validateSingleSuccess = true
		}

		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)

		if !validateManifestSuccess && context.FailFast {
			break
		}
	}

	return validateManifestSuccess, validateManifestErrors
}

// Validate implement Validatable
func (v MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := (len(manifests) == 0 && !context.Negative)
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		validateManifestSuccess, validateManifestErrors := v.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, validateManifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	return validateSuccess, validateErrors
}
