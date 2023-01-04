package validators

import (
	"fmt"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// IsSubsetValidator validate whether value of Path contains Content
type IsSubsetValidator struct {
	Path    string
	Content interface{}
}

func (v IsSubsetValidator) failInfo(actual interface{}, index int, not bool) []string {
	expectedYAML := common.TrustedMarshalYAML(v.Content)
	actualYAML := common.TrustedMarshalYAML(actual)

	log.WithField("validator", "is_subset").Debugln("expected content:", expectedYAML)
	log.WithField("validator", "is_subset").Debugln("actual content:", actualYAML)

	return splitInfof(
		setFailFormat(not, true, true, false, " to contain"),
		index,
		v.Path,
		expectedYAML,
		actualYAML,
	)
}

// Validate implement Validatable
func (v IsSubsetValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		actualMap, actualOk := actual.(map[interface{}]interface{})
		contentMap, contentOk := v.Content.(map[interface{}]interface{})

		if actualOk && contentOk {
			found := validateSubset(actualMap, contentMap)

			if found == context.Negative {
				validateSuccess = false
				errorMessage := v.failInfo(actual, idx, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}

			validateSuccess = determineSuccess(idx, validateSuccess, true)
			continue
		}

		actualYAML, _ := yaml.Marshal(actual)
		validateSuccess = false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect '%s' to be an object, got:\n%s",
			v.Path,
			string(actualYAML),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
