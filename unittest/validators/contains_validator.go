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
func (a ContainsValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if actual, ok := actual.([]interface{}); ok {
		found := false
		for _, ele := range actual {
			if reflect.DeepEqual(ele, a.Content) {
				found = true
			}
		}
		if found != not {
			return true, []string{}
		}
		return false, a.failInfo(actual, not)
	}

	actualYAML, _ := yaml.Marshal(actual)
	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be an array, got:\n%s",
		a.Path,
		string(actualYAML),
	))
}
