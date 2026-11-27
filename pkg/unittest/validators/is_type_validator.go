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
	Path         string
	Type         string
	ParseOptions `mapstructure:",squash"`
}

func (itv IsTypeValidator) failInfo(actual string, manifestIndex, valueIndex int, not bool) []string {
	customMessage := " to be of type"

	log.WithField("validator", "is_type").Debugln("expected type:", itv.Type)
	log.WithField("validator", "is_type").Debugln("actual type:", actual)

	return splitInfof(
		setFailFormat(not, true, true, false, customMessage),
		manifestIndex,
		valueIndex,
		itv.Path,
		itv.Type,
		actual,
	)
}

func (itv IsTypeValidator) validateManifest(manifest common.K8sManifest, manifestIndex int, context *ValidateContext) (bool, []string) {
	actuals, err := valueutils.GetValueOfSetPath(manifest, itv.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	actuals, err = itv.resolveActuals(actuals, itv.Path)
	if err != nil {
		return false, splitInfof(errorFormat, manifestIndex, -1, err.Error())
	}

	if len(actuals) == 0 && !context.Negative {
		return false, splitInfof(errorFormat, manifestIndex, -1,
			fmt.Sprintf("unknown path %s", itv.describePath(itv.Path)))
	}

	manifestSuccess := (len(actuals) == 0 && context.Negative)
	var manifestErrors []string

	for actualIndex, actual := range actuals {
		singleSuccess := false
		var (
			singleErrors []string
			actualType   string
		)
		// reflect.TypeOf returns nil for a nil interface, which is the
		// case for a YAML key that exists but is unset (e.g. `foo:` with
		// no value). Preserve that as the empty string so the type
		// comparison below cleanly fails the assertion instead of
		// panicking in String() (#794).
		if rtoa := reflect.TypeOf(actual); rtoa != nil {
			actualType = rtoa.String()
		}

		if (actualType == itv.Type) == context.Negative {
			singleErrors = itv.failInfo(actualType, manifestIndex, actualIndex, context.Negative)
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
func (itv IsTypeValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		manifestSuccess, manifestErrors := itv.validateManifest(manifest, idx, context)
		validateErrors = append(validateErrors, manifestErrors...)
		validateSuccess = determineSuccess(idx, validateSuccess, manifestSuccess)

		if !validateSuccess && context.FailFast {
			break
		}
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := itv.failInfo("no manifest found", -1, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
