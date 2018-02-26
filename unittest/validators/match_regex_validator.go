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
func (a MatchRegexValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest := context.Docs[context.Index]

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	p, err := regexp.Compile(a.Pattern)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if s, ok := actual.(string); ok {
		if p.MatchString(s) != context.Negative {
			return true, []string{}
		}
		return false, a.failInfo(s, context.Negative)
	}

	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be a string, got:\n%s",
		a.Path,
		common.TrustedMarshalYAML(actual),
	))
}
