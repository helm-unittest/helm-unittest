package validators

import (
	"github.com/lrills/helm-unittest/internal/common"
	log "github.com/sirupsen/logrus"
)

// IsKindValidator validate kind of manifest is Of
type IsKindValidator struct {
	Of string
}

func (v IsKindValidator) failInfo(actual interface{}, index int, not bool) []string {
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to be kind"

	log.WithField("validator", "is_kind").Debugln("expected content:", v.Of)
	log.WithField("validator", "is_kind").Debugln("actual content:", actualYAML)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			index,
			v.Of,
		)
	}
	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		index,
		v.Of,
		actualYAML,
	)
}

// Validate implement Validatable
func (v IsKindValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		if kind, ok := manifest["kind"].(string); (ok && kind == v.Of) == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(manifest["kind"], idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
