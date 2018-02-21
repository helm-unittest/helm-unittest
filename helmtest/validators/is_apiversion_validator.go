package validators

import "github.com/lrills/helm-test/helmtest/common"

// IsAPIVersionValidator validate apiVersion of manifest is Of
type IsAPIVersionValidator struct {
	Of string
}

func (a IsAPIVersionValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	isAPIVersionFailFormat := "Expected" + notAnnotation + " apiVersion:%s"
	if not {
		return splitInfof(isAPIVersionFailFormat, a.Of)
	}
	return splitInfof(isAPIVersionFailFormat+"\nActual:%s", a.Of, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (a IsAPIVersionValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if kind, ok := manifest["apiVersion"].(string); (ok && kind == a.Of) != not {
		return true, []string{}
	}
	return false, a.failInfo(manifest["apiVersion"], not)
}
