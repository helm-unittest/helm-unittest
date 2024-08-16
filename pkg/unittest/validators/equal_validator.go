package validators

import (
	"encoding/base64"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// EqualValidator validate whether the value of Path equal to Value
type EqualValidator struct {
	Path         string
	Value        interface{}
	DecodeBase64 bool `yaml:"decodeBase64"`
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
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		validateSuccess, validateErrors = a.validateManifest(idx, manifest, context, validateSuccess, validateErrors)
	}

	return validateSuccess, validateErrors
}

func (a EqualValidator) validateManifest(idx int, manifest common.K8sManifest, context *ValidateContext, validateSuccess bool, validateErrors []string) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		validateSuccess = false
		errorMessage := splitInfof(errorFormat, idx, err.Error())
		validateErrors = append(validateErrors, errorMessage...)
		return validateSuccess, validateErrors
	}

	if len(actual) == 0 {
		validateSuccess = false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown path %s", a.Path))
		validateErrors = append(validateErrors, errorMessage...)
		return validateSuccess, validateErrors
	}

	singleActual := actual[0]
	if s, ok := singleActual.(string); ok {
		if a.DecodeBase64 {
			decodedSingleActual, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				validateSuccess = false
				errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unable to decode base64 expected content %s", singleActual))
				validateErrors = append(validateErrors, errorMessage...)
				return validateSuccess, validateErrors
			}
			s = string(decodedSingleActual)
		}
		singleActual = uniformContent(s)
	}

	if reflect.DeepEqual(a.Value, singleActual) == context.Negative {
		validateSuccess = false
		errorMessage := a.failInfo(singleActual, idx, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
		return validateSuccess, validateErrors
	}

	validateSuccess = determineSuccess(idx, validateSuccess, true)

	return validateSuccess, validateErrors
}
