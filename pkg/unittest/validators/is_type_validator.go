package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// IsTypeValidator validate the type of the value at Path
type IsTypeValidator struct {
	Path string
	Type string
}

func (t IsTypeValidator) failInfo(actual string, manifestIndex, valueIndex int, not bool) []string {
	customMessage := " to be of type"

	log.WithField("validator", "is_type").Debugln("expected type:", t.Type)
	log.WithField("validator", "is_type").Debugln("actual type:", actual)

	return splitInfof(
		setFailFormat(not, true, true, false, customMessage),
		manifestIndex,
		valueIndex,
		t.Path,
		t.Type,
		actual,
	)
}

func (t IsTypeValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, t.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1, fmt.Sprintf("unknown path %s", t.Path))
	}

	manifestSuccess := (len(actuals) == 0 && context.Negative)
	var manifestErrors []string

	for actualIndex, actual := range actuals {
		singleSuccess := false
		var singleErrors []string
		actualType := reflect.TypeOf(actual).String()

		if (actualType == t.Type) == context.Negative {
			singleErrors = t.failInfo(actualType, manifestIndex, actualIndex, context.Negative)
		} else {
			singleSuccess = true
		}

		manifestErrors = append(manifestErrors, singleErrors...)
		manifestSuccess = determineSuccess(actualIndex, manifestSuccess, singleSuccess)

		if !manifestSuccess && context.FailFast {
			break
		}
	}

	return manifestSuccess, manifestErrors
}

// Validate implement Validatable
func (t IsTypeValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		manifestSuccess, manifestErrors := t.validateManifest(manifest, idx, context)
		validateErrors = append(validateErrors, manifestErrors...)
		validateSuccess = determineSuccess(idx, validateSuccess, manifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := t.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
