package validators

import (
	"fmt"
	"reflect"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
	yaml "gopkg.in/yaml.v2"
)

// ContainsValidator validate whether value of Path is an array and contains Content
type ContainsValidator struct {
	Path    string
	Content interface{}
}

func (a ContainsValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}
	containsFailFormat := `
Path:%s
Expected` + notAnnotation + ` to contain:
%s
Actual:
%s
`
	return splitInfof(
		containsFailFormat,
		a.Path,
		common.TrustedMarshalYAML([]interface{}{a.Content}),
		common.TrustedMarshalYAML(actual),
	)
}

// Validate implement Validatable
func (a ContainsValidator) Validate(context *ValidateContext) (bool, []string) {
	manifest := context.Docs[context.Index]

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if actual, ok := actual.([]interface{}); ok {
		found := false
		for _, ele := range actual {
			if reflect.DeepEqual(ele, a.Content) {
				found = true
			}
		}
		if found != context.Negative {
			return true, []string{}
		}
		return false, a.failInfo(actual, context.Negative)
	}

	actualYAML, _ := yaml.Marshal(actual)
	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be an array, got:\n%s",
		a.Path,
		string(actualYAML),
	))
}
