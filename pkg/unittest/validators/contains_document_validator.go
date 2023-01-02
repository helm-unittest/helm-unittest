package validators

import (
	"fmt"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/valueutils"
)

// IsSubsetValidator validate whether value of Path contains Content
type ContainsDocumentValidator struct {
	Kind       string
	APIVersion string
	Name       string
	Namespace  string
}

func (v ContainsDocumentValidator) failInfo(actual interface{}, index int, not bool) []string {

	return splitInfof(
		setFailFormat(not, false, false, false, " to contain document"),
		index,
		fmt.Sprintf("Kind = %s, apiVersion = %s", v.Kind, v.APIVersion),
	)
}

func (v ContainsDocumentValidator) validateManifest(manifest common.K8sManifest, negative bool) bool {
	if kind, ok := manifest["kind"].(string); (ok && kind == v.Kind) == negative {
		// if no match, move onto next document
		return false
	}

	if api, ok := manifest["apiVersion"].(string); (ok && api == v.APIVersion) == negative {
		// if no match, move onto next document
		return false
	}

	if v.Name != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.name")
		if err != nil {
			// fail on not found match
			return false
		}

		if (actual == v.Name) == negative {
			return false
		}
	}

	if v.Namespace != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.namespace")
		if err != nil {
			// fail on not found match
			return false
		}

		if (actual == v.Namespace) == negative {
			return false
		}
	}

	return true
}

// Validate implement Validatable
func (v ContainsDocumentValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
		validateSuccess = v.validateManifest(manifest, context.Negative)
		if validateSuccess {
			break
		}
		continue
	}
	if !validateSuccess {
		errorMessage := v.failInfo(v.Kind, 0, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	}
	validateSuccess = determineSuccess(1, validateSuccess, true)

	return validateSuccess, validateErrors
}
