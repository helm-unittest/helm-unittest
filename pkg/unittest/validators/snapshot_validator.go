package validators

import (
	"fmt"
	"strconv"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

const (
	pathValidation          snapshotValidation = "path"
	matchRegexValidation    snapshotValidation = "matchRegex"
	notMatchRegexValidation snapshotValidation = "notMatchRegex"
)

type snapshotValidation string

// MatchSnapshotValidator validate snapshot of value of Path the same as cached
type MatchSnapshotValidator struct {
	Path          string
	MatchRegex    *MatchRegex
	NotMatchRegex *NotMatchRegex
}

type MatchRegex struct {
	Pattern string
}

type NotMatchRegex struct {
	Pattern string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, manifestIndex, actualIndex int, not bool, val snapshotValidation) []string {
	log.WithField("validator", "snapshot").Debugln("expected content:", compared.CachedSnapshot)
	log.WithField("validator", "snapshot").Debugln("actual content:", compared.NewSnapshot)

	var result []string

	if val == pathValidation {
		msg := fmt.Sprintf(" to match snapshot %s", strconv.Itoa(int(compared.Index)))
		var infoToShow string
		if not {
			infoToShow = compared.CachedSnapshot
		} else {
			infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
		}
		result = splitInfof(
			setFailFormat(not, true, false, false, msg),
			manifestIndex,
			actualIndex,
			v.Path,
			infoToShow,
		)
	} else if val == matchRegexValidation {
		msg := fmt.Sprintf(" pattern '%s' not found in snapshot", v.MatchRegex.Pattern)
		result = splitInfof(
			setFailFormat(not, false, false, false, msg),
			manifestIndex,
			actualIndex,
			compared.CachedSnapshot,
		)
	} else if val == notMatchRegexValidation {
		msg := fmt.Sprintf(" pattern '%s' should not be in snapshot", v.NotMatchRegex.Pattern)
		result = splitInfof(
			setFailFormat(not, false, false, false, msg),
			manifestIndex,
			actualIndex,
			compared.CachedSnapshot,
		)
	}
	return result
}

func (v MatchSnapshotValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	validateManifestSuccess := len(actual) == 0 && context.Negative
	var validateManifestErrors []string

	for actualIndex, singleActual := range actual {
		validateSingleSuccess := false
		var validateSingleErrors []string
		result := context.CompareToSnapshot(singleActual)

		if result.Passed == context.Negative {
			validateSingleErrors = v.failInfo(result, manifestIndex, actualIndex, context.Negative, pathValidation)
		} else {
			validateSingleSuccess = true
		}

		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)

		if !validateManifestSuccess && context.FailFast {
			break
		}

		if validateManifestSuccess && v.MatchRegex != nil {
			if matches, err := valueutils.MatchesPattern(manifest.ToString(), v.MatchRegex.Pattern); err != nil {
				return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("failed to compile regex pattern %s: %v", v.MatchRegex.Pattern, err))
			} else if !matches {
				validateManifestSuccess = false
				result.Passed = false
				validateSingleErrors = v.failInfo(result, manifestIndex, actualIndex, context.Negative, matchRegexValidation)
				validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
			}
		}

		if validateManifestSuccess && v.NotMatchRegex != nil {
			if matches, err := valueutils.MatchesPattern(manifest.ToString(), v.NotMatchRegex.Pattern); err != nil {
				return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("failed to compile regex pattern %s: %v", v.MatchRegex.Pattern, err))
			} else if matches {
				validateManifestSuccess = false
				result.Passed = false
				validateSingleErrors = v.failInfo(result, manifestIndex, actualIndex, context.Negative, notMatchRegexValidation)
				validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
			}
		}
	}
	return validateManifestSuccess, validateManifestErrors
}

// Validate implement Validatable
func (v MatchSnapshotValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := len(manifests) == 0 && !context.Negative
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
