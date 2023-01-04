package validators

import (
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/lrills/helm-unittest/internal/common"
)

// EqualRawValidator validate whether the raw value equal to Value
type EqualRawValidator struct {
	Value string
}

func (a EqualRawValidator) failInfo(actual interface{}, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(a.Value)
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to equal"

	log.WithField("validator", "equal_raw").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "equal_raw").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			-1,
			expectedYAML,
		)
	}

	return splitInfof(
		setFailFormat(not, false, true, true, customMessage),
		-1,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

// Validate implement Validatable
func (a EqualRawValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual := uniformContent(manifest[common.RAW])

		if reflect.DeepEqual(a.Value, actual) == context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo(actual, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
