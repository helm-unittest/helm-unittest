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
func (a IsNullValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if actual == nil != not {
		return true, []string{}
	}
	return false, a.failInfo(actual, not)
}
