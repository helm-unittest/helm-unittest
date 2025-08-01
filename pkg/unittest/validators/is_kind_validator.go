package validators

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	log "github.com/sirupsen/logrus"
)

// IsKindValidator validate kind of manifest is Of
type IsKindValidator struct {
	Of string
}

func (v IsKindValidator) failInfo(actual any, manifestIndex, actualIndex int, not bool) []string {
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to be kind"

	log.WithField("validator", "is_kind").Debugln("expected content:", v.Of)
	log.WithField("validator", "is_kind").Debugln("actual content:", actualYAML)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			manifestIndex,
			actualIndex,
			v.Of,
		)
	}
	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		manifestIndex,
		actualIndex,
		v.Of,
		actualYAML,
	)
}

// Validate implement Validatable
func (v IsKindValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		if kind, ok := manifest["kind"].(string); (ok && kind == v.Of) == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(manifest["kind"], manifestIndex, -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		validateSuccess = determineSuccess(manifestIndex, validateSuccess, true)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
