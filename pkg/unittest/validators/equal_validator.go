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

func (a EqualValidator) failInfo(actual interface{}, manifestIndex, actualIndex int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(a.Value)
	actualYAML := common.TrustedMarshalYAML(actual)
	customMessage := " to equal"

	log.WithField("validator", "equal").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "equal").Debugln("actual content:", actual)

	if not {
		return splitInfof(
			setFailFormat(not, true, false, false, customMessage),
			manifestIndex,
			actualIndex,
			a.Path,
			expectedYAML,
		)
	}

	return splitInfof(
		setFailFormat(not, true, true, true, customMessage),
		manifestIndex,
		actualIndex,
		a.Path,
		expectedYAML,
		actualYAML,
		diff(expectedYAML, actualYAML),
	)
}

func (a EqualValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", a.Path))
	}

	validateManifestSuccess := (len(actuals) == 0 && context.Negative)
	var validateManifestErrors []string

	for actualIndex, actual := range actuals {
		validateSingleSuccess, validateSingleErrors := a.validateSingleActual(actual, manifestIndex, actualIndex, context)
		validateManifestErrors = append(validateManifestErrors, validateSingleErrors...)
		validateManifestSuccess = determineSuccess(actualIndex, validateManifestSuccess, validateSingleSuccess)
		if !validateSingleSuccess && context.FailFast {
			break
		}
	}

	return validateManifestSuccess, validateManifestErrors
}

func (a EqualValidator) validateSingleActual(actual interface{}, manifestIndex, actualIndex int, context *ValidateContext) (bool, []string) {
	if s, ok := actual.(string); ok {
		if a.DecodeBase64 {
			decodedSingleActual, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return false, splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf("unable to decode base64 expected content %s", actual))
			}
			s = string(decodedSingleActual)
		}
		actual = uniformContent(s)
	}

	if reflect.DeepEqual(a.Value, actual) == context.Negative {
		return false, a.failInfo(actual, manifestIndex, actualIndex, context.Negative)
	}

	return true, []string{}
}

// Validate implement Validatable
func (a EqualValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		validateManifestSuccess, validateManifestErrors := a.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, validateManifestErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, validateManifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := a.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
