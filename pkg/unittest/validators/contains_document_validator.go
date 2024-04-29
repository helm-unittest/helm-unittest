package validators

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// ContainsDocumentValidator validate whether value of Path contains Content
type ContainsDocumentValidator struct {
	Kind       string
	APIVersion string
	Name       string // optional
	Namespace  string // optional
	Any        bool   // optional
}

func (v ContainsDocumentValidator) failInfo(actual interface{}, index int, not bool) []string {

	log.WithField("validator", "contains_document").Debugln("index content:", index)
	log.WithField("validator", "contains_document").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, false, false, false, " to contain document"),
		index,
		v.joinOutput(),
	)
}

// joinOutput constructs a string representation of the ContainsDocumentValidator
// object with the provided fields: Kind, apiVersion, Name, and Namespace.
func (v ContainsDocumentValidator) joinOutput() string {
	parts := []string{
		fmt.Sprintf("Kind = %s, apiVersion = %s", v.Kind, v.APIVersion),
	}
	if v.Name != "" {
		parts = append(parts, fmt.Sprintf("Name = %s", v.Name))
	}
	if v.Namespace != "" {
		parts = append(parts, fmt.Sprintf("Namespace = %s", v.Namespace))
	}
	return strings.Join(parts, ", ")
}

func (v ContainsDocumentValidator) validateManifest(manifest common.K8sManifest) bool {
	if kind, ok := manifest["kind"].(string); ok && kind != v.Kind {
		// if no match, move onto next document
		return false
	}

	if api, ok := manifest["apiVersion"].(string); ok && api != v.APIVersion {
		// if no match, move onto next document
		return false
	}

	if v.Name != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.name")
		if err != nil {
			// fail on not found match
			return false
		}

		if len(actual) == 0 || actual[0] != v.Name {
			return false
		}
	}

	if v.Namespace != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.namespace")
		if err != nil {
			// fail on not found match
			log.WithField("validator", "contains_document").Debugln("error:", err)
			return false
		}

		log.WithField("validator", "contains_document").Debugln("namespace path:", actual)
		if len(actual) == 0 || actual[0] != v.Namespace {
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

	for idx, manifest := range manifests {
		singleSuccess := v.validateManifest(manifest)

		if singleSuccess == context.Negative {
			singleSuccess = false
			errorMessage := v.failInfo(v.Kind, idx, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
		} else {
			singleSuccess = true
			if v.Any != context.Negative {
				validateSuccess = true
				validateErrors = []string{}
				// Stop searching as we already found a successful match.
				break
			}
		}

		validateSuccess = determineSuccess(idx, validateSuccess, singleSuccess)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := v.failInfo(v.Kind, 0, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
