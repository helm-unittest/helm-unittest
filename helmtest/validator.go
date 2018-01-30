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
	Validate(docs []K8sManifest, idx int) (bool, string)
}

func printFailf(format string, replacements ...string) string {
	intentedFormat := strings.Trim(strings.Replace(format, "\n", "\n\t", -1), "\t")
	indentedReplacements := make([]interface{}, len(replacements))
	for i, r := range replacements {
		indentedReplacements[i] = strings.Trim(strings.Replace(r, "\n", "\n\t\t", -1), "\n\t ")
	}
	return fmt.Sprintf(intentedFormat, indentedReplacements...)
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

const equalFailFormat = `
Path: %s
Expected:
	%s
Actual:
	%s
Diff:
	%s
`

func (a EqualValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}

	if !reflect.DeepEqual(a.Value, actual) {
		expectedYAML := trustedMarshalYAML(a.Value)
		actualYAML := trustedMarshalYAML(actual)
		return false, printFailf(
			equalFailFormat,
			a.Path,
			expectedYAML,
			actualYAML,
			diff(expectedYAML, actualYAML),
		)
	}
	return true, ""
}

type MatchRegexValidator struct {
	Path    string
	Pattern string
}

const regexFailFormat = `
Path: %s
Expected to Match: %s
Actual: %s
`

func (a MatchRegexValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}

	p, err := regexp.Compile(a.Pattern)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}

	if s, ok := actual.(string); ok {
		if p.MatchString(s) {
			return true, ""
		}
		return false, printFailf(regexFailFormat, a.Path, a.Pattern, s)
	}
	return false, printFailf(errorFormat, fmt.Sprintf(
		"expect '%s' to be a string, got:\n%s",
		a.Path,
		trustedMarshalYAML(actual),
	))
}

const containsFailFormat = `
Path: %s
Expected Contains:
	%s
Actual:
	%s
`

type ContainsValidator struct {
	Path    string
	Content interface{}
}

func (a ContainsValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}
	if actual, ok := actual.([]interface{}); ok {
		for _, ele := range actual {
			if reflect.DeepEqual(ele, a.Content) {
				return true, ""
			}
		}

		return false, printFailf(
			containsFailFormat,
			a.Path,
			trustedMarshalYAML([]interface{}{a.Content}),
			trustedMarshalYAML(actual),
		)
	}
	actualYAML, _ := yaml.Marshal(actual)
	return false, printFailf(errorFormat, fmt.Sprintf(
		"expect '%s' to be an array, got:\n%s",
		a.Path,
		string(actualYAML),
	))
}

type IsNullValidator struct {
	Path string
}

const isNullFailFormat = `
Path: %s
Expected: null
Actual:
	%s
`

func (a IsNullValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}

	if actual == nil {
		return true, ""
	}
	return false, printFailf(isNullFailFormat, a.Path, trustedMarshalYAML(actual))
}

type IsEmptyValidator struct {
	Path string
}

const isEmptyFailFormat = `
Path: %s
Expected to be empty, got:
	%s
`

func (a IsEmptyValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	actual, err := GetValueOfSetPath(docs[idx], a.Path)
	if err != nil {
		return false, printFailf(errorFormat, err.Error())
	}

	if actual == nil || actual == reflect.Zero(reflect.TypeOf(actual)).Interface() {
		return true, ""
	}
	return false, printFailf(isEmptyFailFormat, a.Path, trustedMarshalYAML(actual))
}

type IsKindValidator struct {
	Of string
}

const isKindFailFormat = `
Expected 'kind': %s
Actual: %s
`

func (a IsKindValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	if kind, ok := docs[idx]["kind"].(string); ok && kind == a.Of {
		return true, ""
	}
	return false, printFailf(isKindFailFormat, a.Of, trustedMarshalYAML(docs[idx]["kind"]))
}

type IsAPIVersionValidator struct {
	Of string
}

const isAPIVersionFailFormat = `
Expected 'apiVersion': %s
Actual: %s
`

func (a IsAPIVersionValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	if kind, ok := docs[idx]["apiVersion"].(string); ok && kind == a.Of {
		return true, ""
	}
	return false, printFailf(isAPIVersionFailFormat, a.Of, trustedMarshalYAML(docs[idx]["apiVersion"]))
}

type HasDocumentsValidator struct {
	Count int
}

const hasDocumentsFailFormat = `
Expected: %s
Actual: %s
`

func (a HasDocumentsValidator) Validate(docs []K8sManifest, idx int) (bool, string) {
	if len(docs) == a.Count {
		return true, ""
	}
	return false, printFailf(hasDocumentsFailFormat, strconv.Itoa(a.Count), strconv.Itoa(len(docs)))
}
