package helmtest

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v2"
)

type Validatable interface {
	Validate(docs []K8sManifest, idx int, not bool) (bool, []string)
}

func splitInfof(format string, replacements ...string) []string {
	intentedFormat := strings.Trim(format, "\t\n ")
	indentedReplacements := make([]interface{}, len(replacements))
	for i, r := range replacements {
		indentedReplacements[i] = "\t" + strings.Trim(
			strings.Replace(r, "\n", "\n\t", -1),
			"\n\t ",
		)
	}
	return strings.Split(
		fmt.Sprintf(intentedFormat, indentedReplacements...),
		"\n",
	)
}

func diff(expected string, actual string) string {
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})
	return diff
}

const errorFormat = `
Error:
%s
`

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

	expectedYAML := trustedMarshalYAML(a.Value)
	if !not {
		actualYAML := trustedMarshalYAML(actual)
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

func (a EqualValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}
	if reflect.DeepEqual(a.Value, actual) == not {
		return false, a.failInfo(actual, not)
	}
	return true, []string{}
}

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

func (a MatchRegexValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	p, err := regexp.Compile(a.Pattern)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if s, ok := actual.(string); ok {
		if p.MatchString(s) != not {
			return true, []string{}
		}
		return false, a.failInfo(s, not)
	}
	return false, splitInfof(errorFormat, fmt.Sprintf(
		"expect '%s' to be a string, got:\n%s",
		a.Path,
		trustedMarshalYAML(actual),
	))
}

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
		trustedMarshalYAML([]interface{}{a.Content}),
		trustedMarshalYAML(actual),
	)
}

func (a ContainsValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
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
	return splitInfof(isNullFailFormat, a.Path, trustedMarshalYAML(actual))
}

func (a IsNullValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	if actual == nil != not {
		return true, []string{}
	}
	return false, a.failInfo(actual, not)
}

type IsEmptyValidator struct {
	Path string
}

func (a IsEmptyValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isEmptyFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be empty, got:
%s
`
	return splitInfof(isEmptyFailFormat, a.Path, trustedMarshalYAML(actual))
}

func (a IsEmptyValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	actualValue := reflect.ValueOf(actual)
	var isEmpty bool
	switch actualValue.Kind() {
	case reflect.Invalid:
		isEmpty = true
	case reflect.Array, reflect.Map, reflect.Slice:
		isEmpty = actualValue.Len() == 0
	default:
		zero := reflect.Zero(actualValue.Type())
		isEmpty = reflect.DeepEqual(actual, zero.Interface())
	}

	if isEmpty != not {
		return true, []string{}
	}
	return false, a.failInfo(actual, not)
}

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
	return splitInfof(isKindFailFormat+"\nActual:%s", a.Of, trustedMarshalYAML(actual))
}

func (a IsKindValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	if kind, ok := docs[idx]["kind"].(string); (ok && kind == a.Of) != not {
		return true, []string{}
	}
	return false, a.failInfo(docs[idx]["kind"], not)
}

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
	return splitInfof(isAPIVersionFailFormat+"\nActual:%s", a.Of, trustedMarshalYAML(actual))
}

func (a IsAPIVersionValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	if kind, ok := docs[idx]["apiVersion"].(string); (ok && kind == a.Of) != not {
		return true, []string{}
	}
	return false, a.failInfo(docs[idx]["apiVersion"], not)
}

type HasDocumentsValidator struct {
	Count int
}

func (a HasDocumentsValidator) failInfo(actual int, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	hasDocumentsFailFormat := "Expected documents count" + notAnnotation + ":%s"
	if not {
		return splitInfof(hasDocumentsFailFormat, strconv.Itoa(a.Count))
	}
	return splitInfof(
		hasDocumentsFailFormat+"\nActual:%s",
		strconv.Itoa(a.Count),
		strconv.Itoa(actual),
	)
}

func (a HasDocumentsValidator) Validate(docs []K8sManifest, idx int, not bool) (bool, []string) {
	if len(docs) == a.Count != not {
		return true, []string{}
	}
	return false, a.failInfo(len(docs), not)
}
