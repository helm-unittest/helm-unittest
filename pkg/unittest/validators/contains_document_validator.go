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

func (cdv ContainsDocumentValidator) failInfo(actual any, manifestIndex, assertIndex int, not bool) []string {

	log.WithField("validator", "contains_document").Debugln("index content:", manifestIndex)
	log.WithField("validator", "contains_document").Debugln("actual content:", actual)

	return splitInfof(
		setFailFormat(not, false, false, false, " to contain document"),
		manifestIndex,
		assertIndex,
		cdv.joinOutput(),
	)
}

// joinOutput constructs a string representation of the ContainsDocumentValidator
// object with the provided fields: Kind, apiVersion, Name, and Namespace.
func (cdv ContainsDocumentValidator) joinOutput() string {
	parts := []string{
		fmt.Sprintf("Kind = %s, apiVersion = %s", cdv.Kind, cdv.APIVersion),
	}
	if cdv.Name != "" {
		parts = append(parts, fmt.Sprintf("Name = %s", cdv.Name))
	}
	if cdv.Namespace != "" {
		parts = append(parts, fmt.Sprintf("Namespace = %s", cdv.Namespace))
	}
	return strings.Join(parts, ", ")
}

func (cdv ContainsDocumentValidator) validateManifest(manifest common.K8sManifest) bool {
	if !cdv.matchKind(manifest) {
		return false
	}

	if !cdv.matchAPIVersion(manifest) {
		return false
	}

	if !cdv.matchName(manifest) {
		return false
	}

	if !cdv.matchNamespace(manifest) {
		return false
	}

	return true
}

func (cdv ContainsDocumentValidator) matchKind(manifest common.K8sManifest) bool {
	if kind, ok := manifest["kind"].(string); ok && kind != cdv.Kind {
		// if no match, move onto next document
		return false
	}
	return true
}

func (cdv ContainsDocumentValidator) matchAPIVersion(manifest common.K8sManifest) bool {
	if api, ok := manifest["apiVersion"].(string); ok && api != cdv.APIVersion {
		// if no match, move onto next document
		return false
	}
	return true
}

func (cdv ContainsDocumentValidator) matchName(manifest common.K8sManifest) bool {
	if cdv.Name != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.name")
		if err != nil {
			// fail on not found match
			return false
		}

		if len(actual) == 0 || actual[0] != cdv.Name {
			return false
		}
	}
	return true
}

func (cdv ContainsDocumentValidator) matchNamespace(manifest common.K8sManifest) bool {
	if cdv.Namespace != "" {
		actual, err := valueutils.GetValueOfSetPath(manifest, "metadata.namespace")
		if err != nil {
			// fail on not found match
			log.WithField("validator", "contains_document").Debugln("error:", err)
			return false
		}

		log.WithField("validator", "contains_document").Debugln("namespace path:", actual)
		if len(actual) == 0 || actual[0] != cdv.Namespace {
			return false
		}
	}
	return true
}

// Validate implement Validatable
func (cdv ContainsDocumentValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests := context.getManifests()

	validateSuccess := false
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		manifestSuccess := cdv.validateManifest(manifest)

		if manifestSuccess == context.Negative {
			manifestSuccess = false
			errorMessage := cdv.failInfo(cdv.Kind, idx, -1, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)

			if context.FailFast {
				// Stop searching as we already found a failing match.
				validateSuccess = false
				break
			}
		} else {
			manifestSuccess = true
			if cdv.Any != context.Negative {
				validateSuccess = true
				validateErrors = []string{}
				// Stop searching as we already found a successful match.
				break
			}
		}

		validateSuccess = determineSuccess(idx, validateSuccess, manifestSuccess)
	}

	if len(manifests) == 0 && !context.Negative {
		errorMessage := cdv.failInfo(cdv.Kind, 0, -1, context.Negative)
		validateErrors = append(validateErrors, errorMessage...)
	} else if len(manifests) == 0 && context.Negative {
		validateSuccess = true
	}

	return validateSuccess, validateErrors
}
