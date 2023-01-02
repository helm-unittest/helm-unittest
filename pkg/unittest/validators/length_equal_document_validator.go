package validators

import (
	"fmt"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// LengthEqualDocumentsValidator validate whether the count of manifests rendered form template is Count
type LengthEqualDocumentsValidator struct {
	Paths []string // optional
	Path  string   // optional
	Count int      // optional if paths defined
}

func (v LengthEqualDocumentsValidator) failInfo(actual int, not bool) []string {
	expectedCount := strconv.Itoa(v.Count)
	actualCount := strconv.Itoa(actual)
	customMessage := " documents count to be"

	log.WithField("validator", "length_equal").Debugln("expected content:", expectedCount)
	log.WithField("validator", "length_equal").Debugln("actual content:", actualCount)

	if not {
		return splitInfof(
			setFailFormat(not, false, false, false, customMessage),
			-1,
			expectedCount,
		)
	}
	return splitInfof(
		setFailFormat(not, false, true, false, customMessage),
		-1,
		expectedCount,
		actualCount,
	)
}

// Validate implement Validatable
func (v LengthEqualDocumentsValidator) Validate(context *ValidateContext) (bool, []string) {
	if len(v.Path) > 0 && v.Count == 0 {
		return false, splitInfof(errorFormat, -1, "'count' field must be set if 'path' is used")
	}
	if len(v.Path) > 0 && len(v.Paths) > 0 {
		return false, splitInfof(errorFormat, -1, "'paths' couldn't be used with 'path'")
	}
	singleMode := len(v.Path) > 0
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}
	validateSuccess := false
	validateErrors := make([]string, 0)
	for idx, manifest := range manifests {
		if singleMode {
			spec, err := valueutils.GetValueOfSetPath(manifest, v.Path)
			if err != nil {
				validateSuccess = false
				errorMessage := splitInfof(errorFormat, idx, err.Error())
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}
			specArr, ok := spec.([]interface{})
			if !ok {
				validateSuccess = false
				errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("%s is not array", v.Path))
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}
			specLen := len(specArr)
			if specLen != v.Count {
				validateSuccess = false
				errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
					"count doesn't match. expected: %d != %d actual", v.Count, specLen))
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}
		} else {
			px := map[string]int{}
			c := true
			for _, p := range v.Paths {
				sp, err := valueutils.GetValueOfSetPath(manifest, p)
				if err != nil {
					validateSuccess = false
					errorMessage := splitInfof(errorFormat, idx, err.Error())
					validateErrors = append(validateErrors, errorMessage...)
					c = false
					break
				}
				specArr, ok := sp.([]interface{})
				if !ok {
					validateSuccess = false
					errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf("%s is not array", p))
					validateErrors = append(validateErrors, errorMessage...)
					c = false
					break
				}
				px[p] = len(specArr)
			}
			if !c {
				continue
			}
			acc := -1
			for k, vv := range px {
				if acc == -1 {
					acc = vv
				} else if acc != vv {
					acc = -1
					validateSuccess = false
					errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
						"%s count is '%d'(doesn't match others)", k, vv))
					validateErrors = append(validateErrors, errorMessage...)
					break
				}
			}
			if acc == -1 {
				continue
			}
		}
		validateSuccess = determineSuccess(idx, validateSuccess, true)
	}
	if !validateSuccess {
		//errorMesasge := v.failInfo(v.Kind, 0, context.Negative)
		errorMessage := []string{}
		validateErrors = append(validateErrors, errorMessage...)
	}
	validateSuccess = determineSuccess(1, validateSuccess, true)
	return validateSuccess, validateErrors
}
