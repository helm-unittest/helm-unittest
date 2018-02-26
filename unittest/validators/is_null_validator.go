package validators

import (
	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// IsNullValidator validate value of Path id kind
type IsNullValidator struct {
	Path string
}

func (a IsNullValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isNullFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be null, got:
%s
`
	return splitInfof(isNullFailFormat, a.Path, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (a IsNullValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest := context.Docs[context.Index]
	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if actual == nil != context.Negative {
		return true, []string{}
	}
	return false, a.failInfo(actual, context.Negative)
}
