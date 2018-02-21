package validators

import (
	"reflect"

	"github.com/lrills/helm-test/helmtest/common"
	"github.com/lrills/helm-test/helmtest/valueutils"
)

// EqualValidator validate whether the value of Path equal to Value
type EqualValidator struct {
	Path  string
	Value interface{}
}

func (a EqualValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to equal"
	}
	failFormat := `
Path:%s
Expected` + notAnnotation + `:
%s`

	expectedYAML := common.TrustedMarshalYAML(a.Value)
	if !not {
		actualYAML := common.TrustedMarshalYAML(actual)
		return splitInfof(
			failFormat+`
Actual:
%s
Diff:
%s
`,
			a.Path,
			expectedYAML,
			actualYAML,
			diff(expectedYAML, actualYAML),
		)
	}
	return splitInfof(failFormat, a.Path, expectedYAML)
}

// Validate implement Validatable
func (a EqualValidator) Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string) {
	manifest, err := assert.GetManifest(docs)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actual, err := valueutils.GetValueOfSetPath(manifest, a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	not := assert.IsNegative()
	if reflect.DeepEqual(a.Value, actual) == not {
		return false, a.failInfo(actual, not)
	}
	return true, []string{}
}
