package validators

import (
	"fmt"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
)

// IsSubsetValidator validate whether value of Path contains Content
type IsSubsetValidator struct {
	Path    string
	Content interface{}
}

func (v IsSubsetValidator) failInfo(actual interface{}, manifestIndex, valueIndex int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(v.Content)
	actualYAML := common.TrustedMarshalYAML(actual)

	log.WithField("validator", "is_subset").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "is_subset").Debugln("actual content:", actualYAML)

	return splitInfof(
		setFailFormat(not, true, true, false, " to contain"),
		manifestIndex,
		valueIndex,
		v.Path,
		expectedYAML,
		actualYAML,
	)
}

func (v IsSubsetValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	manifestValidateSuccess := (len(actual) == 0 && context.Negative)
	var manifestValidateErrors []string

	for actualIndex, singleActual := range actual {
		var errorMessage []string
		actualMap, actualOk := singleActual.(map[string]interface{})
		contentMap, contentOk := v.Content.(map[string]interface{})

		if actualOk && contentOk {
			found := validateSubset(actualMap, contentMap)

			if found == context.Negative {
				errorMessage = v.failInfo(singleActual, manifestIndex, actualIndex, context.Negative)
			}

			manifestValidateErrors = append(manifestValidateErrors, errorMessage...)
			manifestValidateSuccess = determineSuccess(actualIndex, manifestValidateSuccess, found != context.Negative)

			if !manifestValidateSuccess && context.FailFast {
				break
			}

			continue
		}

		actualYAML := common.TrustedMarshalYAML(singleActual)
		errorMessage = splitInfof(errorFormat, manifestIndex, actualIndex, fmt.Sprintf(
			"expect '%s' to be an object, got:\n%s",
			v.Path,
			actualYAML,
		))

		manifestValidateErrors = append(manifestValidateErrors, errorMessage...)
		manifestValidateSuccess = determineSuccess(actualIndex, manifestValidateSuccess, false)
	}

	return manifestValidateSuccess, manifestValidateErrors
}

// Validate implement Validatable
func (v IsSubsetValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		manifestValidateSuccess, manifestValidateErrors := v.validateManifest(manifest, idx, context)
		validateErrors = append(validateErrors, manifestValidateErrors...)
		validateSuccess = determineSuccess(idx, validateSuccess, manifestValidateSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
