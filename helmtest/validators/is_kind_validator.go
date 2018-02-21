package validators

import "github.com/lrills/helm-test/helmtest/common"

// IsKindValidator validate kind of manifest is Of
type IsKindValidator struct {
	Of string
}

func (a IsKindValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	isKindFailFormat := "Expected" + notAnnotation + " kind:%s"
	if not {
		return splitInfof(isKindFailFormat, a.Of)
	}
	return splitInfof(isKindFailFormat+"\nActual:%s", a.Of, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (a IsKindValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if kind, ok := manifest["kind"].(string); (ok && kind == a.Of) != not {
		return true, []string{}
	}
	return false, a.failInfo(manifest["kind"], not)
}
