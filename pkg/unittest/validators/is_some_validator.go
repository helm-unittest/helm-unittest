package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// IsSomeValidator validate value of Path is something aka non-null
type IsSomeValidator struct {
	Path string
}

func (v IsSomeValidator) failInfo(actual interface{}, manifestIndex, actualIndex int, not bool) []string {
	actualYAML := common.TrustedMarshalYAML(actual)

	log.WithField("validator", "is_some").Debugln("actual content:", actualYAML)

	return splitInfof(
		setFailFormat(not, true, false, false, " to be something, got"),
		manifestIndex,
		actualIndex,
		v.Path,
		actualYAML,
	)
}

func (v IsSomeValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actual) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", v.Path))
	}

	manifestSuccess := (len(actual) == 0 && context.Negative)
	var manifestValidateErrors []string

	for actualIndex, singleActual := range actual {
		singleSuccess := false
		var singleValidateErrors []string

		actualValue := reflect.ValueOf(singleActual)
		isSome := actualValue.Kind() != reflect.Invalid

		if isSome == context.Negative {
			singleValidateErrors = v.failInfo(singleActual, manifestIndex, actualIndex, context.Negative)
		} else {
			singleSuccess = true
		}

		manifestValidateErrors = append(manifestValidateErrors, singleValidateErrors...)
		manifestSuccess = determineSuccess(actualIndex, manifestSuccess, singleSuccess)

		if !singleSuccess && context.FailFast {
			break
		}
	}

	return manifestSuccess, manifestValidateErrors
}

// Validate implement Validatable
func (v IsSomeValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for manifestIndex, manifest := range manifests {
		manifestSuccess, manifestValidateErrors := v.validateManifest(manifest, manifestIndex, context)
		validateErrors = append(validateErrors, manifestValidateErrors...)
		validateSuccess = determineSuccess(manifestIndex, validateSuccess, manifestSuccess)

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
