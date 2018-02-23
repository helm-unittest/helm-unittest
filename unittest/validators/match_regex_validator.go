package validators

import (
	"fmt"
	"regexp"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// MatchRegexValidator validate value of Path match Pattern
type MatchRegexValidator struct {
	Path    string
	Pattern string
}

func (a MatchRegexValidator) failInfo(actual string, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	regexFailFormat := `
Path:%s
Expected` + notAnnotation + ` to match:%s
Actual:%s
`
	return splitInfof(regexFailFormat, a.Path, a.Pattern, actual)
}

// Validate implement Validatable
func (a MatchRegexValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	p, err := regexp.Compile(a.Pattern)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if s, ok := actual.(string); ok {
		if p.MatchString(s) != not {
			return true, []string{}
		}
		return false, a.failInfo(s, not)
	}
	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be a string, got:\n%s",
		a.Path,
		common.TrustedMarshalYAML(actual),
	))
}
