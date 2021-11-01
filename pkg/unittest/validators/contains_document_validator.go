package validators

import (
	"fmt"

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

// Validate implement Validatable
func (v ContainsDocumentValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := false
	validateErrors := make([]string, 0)

	fmt.Printf("%+v\n", v)
	for _, manifest := range manifests {
		if kind, ok := manifest["kind"].(string); (ok && kind == v.Kind) == context.Negative {
			// if no match, move onto next document
			continue
		}

		if api, ok := manifest["apiVersion"].(string); (ok && api == v.APIVersion) == context.Negative {
			// if no match, move onto next document
			continue
		}

		if v.Name != "" {
			actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.name")
			if err != nil {
				// fail on not found match
				continue
			}

			if (actual == v.Name) == context.Negative {
				continue
			}
		}

		if v.Namespace != "" {
			actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.namespace")
			if err != nil {
				// fail on not found match
				continue
			}

			if (actual == v.Namespace) == context.Negative {
				continue
			}
		}

		// if we get here the above have held so it is a match
		validateSuccess = true
		break
	}
	if !validateSuccess {
		errorMesasge := v.failInfo(v.Kind, 0, context.Negative)
		validateErrors = append(validateErrors, errorMesasge...)
	}
	validateSuccess = determineSuccess(1, validateSuccess, true)

	return validateSuccess, validateErrors
}
