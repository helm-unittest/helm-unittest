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
	Path          string
	MatchRegex    *MatchRegex
	NotMatchRegex *NotMatchRegex
	// Format specifies the snapshot format: "indexed" (default) or "yaml" (pure YAML with --- separators)
	Format string
}

type MatchRegex struct {
	Pattern string
}

type NotMatchRegex struct {
	Pattern string
}

func (v MatchSnapshotValidator) failInfo(compared *snapshot.CompareResult, manifestIndex, actualIndex int, not bool) []string {
	log.WithField("validator", "snapshot").Debugln("expected content:", compared.CachedSnapshot)
	log.WithField("validator", "snapshot").Debugln("actual content:", compared.NewSnapshot)

	var result []string

	if compared.Msg != "" {
		result = splitInfof(
			setFailFormat(not, false, false, false, compared.Msg),
			manifestIndex,
			actualIndex,
			compared.CachedSnapshot,
		)
	} else {
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
	}
	return result
}

func (v MatchSnapshotValidator) failInfoYAML(compared *snapshot.CompareResult, not bool) []string {
	log.WithField("validator", "snapshot").Debugln("expected content:", compared.CachedSnapshot)
	log.WithField("validator", "snapshot").Debugln("actual content:", compared.NewSnapshot)

	var result []string

	if compared.Msg != "" {
		result = splitInfof(
			setFailFormat(not, false, false, false, compared.Msg),
			-1,
			-1,
			compared.CachedSnapshot,
		)
	} else {
		msg := " to match snapshot (yaml format)"
		var infoToShow string
		if not {
			infoToShow = compared.CachedSnapshot
		} else {
			infoToShow = diff(compared.CachedSnapshot, compared.NewSnapshot)
		}
		result = splitInfof(
			setFailFormat(not, false, false, true, msg),
			-1,
			-1,
			infoToShow,
		)
	}
	return result
}

// isYAMLFormat returns true if the format is set to "yaml"
func (v MatchSnapshotValidator) isYAMLFormat() bool {
	return v.Format == string(snapshot.SnapshotFormatYAML)
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
		withMatchRegex := snapshot.WithMatchRegexPattern("")
		withNotMatchRegex := snapshot.WithNotMatchRegexPattern("")
		if v.MatchRegex != nil && v.MatchRegex.Pattern != "" {
			withMatchRegex = snapshot.WithMatchRegexPattern(v.MatchRegex.Pattern)
		}
		if v.NotMatchRegex != nil && v.NotMatchRegex.Pattern != "" {
			withNotMatchRegex = snapshot.WithNotMatchRegexPattern(v.NotMatchRegex.Pattern)
		}
		result := context.CompareToSnapshot(singleActual, withMatchRegex, withNotMatchRegex)

		if result.Err != nil {
			return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("%v", err))
		}

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

	// Use YAML format if specified
	if v.isYAMLFormat() {
		return v.validateYAMLFormat(manifests, context)
	}

	// Default indexed format
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

// validateYAMLFormat validates all manifests as a single YAML document with --- separators
func (v MatchSnapshotValidator) validateYAMLFormat(manifests []common.K8sManifest, context *ValidateContext) (bool, []string) {
	// Match indexed format behavior: empty manifests pass when not negative
	if len(manifests) == 0 {
		return !context.Negative, []string{}
	}

	// Collect all contents to compare
	var contents []any
	for _, manifest := range manifests {
		if v.Path != "" {
			actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
			if err != nil {
				return false, splitInfof(errorFormat, -1, -1, err.Error())
			}
			for _, a := range actual {
				contents = append(contents, a)
			}
		} else {
			// If no path specified, use the entire manifest
			contents = append(contents, manifest)
		}
	}

	// If path was specified but nothing found, handle like indexed format
	if len(contents) == 0 {
		return context.Negative, []string{}
	}

	// Build options for regex matching
	withMatchRegex := snapshot.WithMatchRegexPattern("")
	withNotMatchRegex := snapshot.WithNotMatchRegexPattern("")
	if v.MatchRegex != nil && v.MatchRegex.Pattern != "" {
		withMatchRegex = snapshot.WithMatchRegexPattern(v.MatchRegex.Pattern)
	}
	if v.NotMatchRegex != nil && v.NotMatchRegex.Pattern != "" {
		withNotMatchRegex = snapshot.WithNotMatchRegexPattern(v.NotMatchRegex.Pattern)
	}

	// Compare all contents as a single YAML document
	result := context.CompareToSnapshotYAML(contents, withMatchRegex, withNotMatchRegex)

	if result.Err != nil {
		return false, splitInfof(errorFormat, -1, -1, result.Err.Error())
	}

	if result.Passed == context.Negative {
		return false, v.failInfoYAML(result, context.Negative)
	}

	return true, nil
}
