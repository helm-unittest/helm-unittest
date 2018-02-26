package validators

import "github.com/lrills/helm-unittest/unittest/common"

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
func (a IsKindValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest := context.Docs[context.Index]

	if kind, ok := manifest["kind"].(string); (ok && kind == a.Of) != context.Negative {
		return true, []string{}
	}
	return false, a.failInfo(manifest["kind"], context.Negative)
}
