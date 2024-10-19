package validators

import (
	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

// IsAPIVersionValidator validate apiVersion of manifest is Of
type IsAPIVersionValidator struct {
	Of string
}

func (v IsAPIVersionValidator) failInfo(actual interface{}, index int, not bool) []string {
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to be apiVersion"

	log.WithField("validator", "is_apiversion").Debugln("expected content:", v.Of)
	log.WithField("validator", "is_apiversion").Debugln("actual content:", actualYAML)

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
func (v IsAPIVersionValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		if kind, ok := manifest["apiVersion"].(string); (ok && kind == v.Of) == context.Negative {
			validateSuccess = false
			errorMessage := v.failInfo(manifest["apiVersion"], idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
