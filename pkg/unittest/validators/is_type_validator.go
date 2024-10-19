package validators

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// IsTypeValidator validate the type of the value at Path
type IsTypeValidator struct {
	Path string
	Type string
}

func (t IsTypeValidator) failInfo(actual string, index int, not bool) []string {
	customMessage := " to be of type"

	log.WithField("validator", "is_type").Debugln("expected type:", t.Type)
	log.WithField("validator", "is_type").Debugln("actual type:", actual)

	return splitInfof(
		setFailFormat(not, true, true, false, customMessage),
		index,
		t.Path,
		t.Type,
		actual,
	)
}

// Validate implement Validatable
func (t IsTypeValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, t.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		if len(actual) == 0 {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("unknown path %s", t.Path))
			validateErrors = append(validateErrors, errorMessage...)
			if context.FailFast {
				break
			}
			continue
		}

		actualType := reflect.TypeOf(actual[0]).String()

		if (actualType == t.Type) == context.Negative {
			validateSuccess = false
			errorMessage := t.failInfo(actualType, idx, context.Negative)
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
