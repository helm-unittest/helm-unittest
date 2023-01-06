package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
)

// EqualValidator validate whether the value of Path equal to Value
type EqualValidator struct {
	Path  string
	Value interface{}
}

func (a EqualValidator) failInfo(actual interface{}, index int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(a.Value)
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to equal"

	log.WithField("validator", "equal").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "equal").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, true, false, false, customMessage),
			index,
			a.Path,
			expectedYAML,
		)
	}

	return splitInfof(
		setFailFormat(not, true, true, true, customMessage),
		index,
		a.Path,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

// Validate implement Validatable
func (a EqualValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if len(actual) == 0 {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown parameter %s", a.Path))
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		singleActual := actual[0]
		if _, ok := singleActual.(string); ok {
			singleActual = uniformContent(singleActual)
		}

		if reflect.DeepEqual(a.Value, singleActual) == context.Negative {
			validateSuccess = false
			errorMessage := a.failInfo(singleActual, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}

	return validateSuccess, validateErrors
}
