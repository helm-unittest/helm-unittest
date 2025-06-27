package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/pmezard/go-difflib/difflib"
)

const errorFormat = `
Error:
%s
`

// For multilines, remove spaces before newlines
// And ensure all decoded content uses LF
var regex = regexp.MustCompile(`(?m)[ ]+\r?\n`)

// SnapshotComparer provide CompareToSnapshot utility to validator
type SnapshotComparer interface {
	CompareToSnapshot(content interface{}, optFns ...func(options *snapshot.CacheOptions) error) *snapshot.CompareResult
}

// ValidateContext the context passed to validators
type ValidateContext struct {
	Docs         []common.K8sManifest
	SelectedDocs *[]common.K8sManifest
	Negative     bool
	SnapshotComparer
	RenderError error
	FailFast    bool
}

func (c *ValidateContext) getManifests() []common.K8sManifest {
	// This here is for making a default for unit tests
	if c.SelectedDocs == nil {
		return c.Docs
	} else {
		return *c.SelectedDocs
	}
}

// Validatable all validators must implement Validate method
type Validatable interface {
	Validate(context *ValidateContext) (bool, []string)
}

// setFailFormat,
// setting the formatting for the failure message.
func setFailFormat(not, path, actual, diff bool, customize string) string {
	var notAnnotation string
	var result string
	if not {
		notAnnotation = " NOT"
	}
	result = `
Expected` + notAnnotation + customize + `:
%s
`
	if path {
		result = `Path:%s` + result
	}
	if actual {
		result += `Actual:
%s
`
	}
	if diff {
		result += `Diff:
%s
`
	}
	return result
}

// splitInfof split multi line string into array of string
func splitInfof(format string, manifestIndex, valuesIndex int, replacements ...string) []string {
	intentedFormat := strings.Trim(format, "\t\n ")
	indentedReplacements := make([]interface{}, len(replacements))
	for i, r := range replacements {
		indentedReplacements[i] = "\t" + strings.Trim(
			strings.Replace(r, "\n", "\n\t", -1),
			"\n\t ",
		)
	}

	splittedStrings := strings.Split(
		fmt.Sprintf(intentedFormat, indentedReplacements...),
		"\n",
	)

	// Only shown multiple values are found for assertions.
	if valuesIndex >= 0 {
		valuesIndexString := []string{fmt.Sprintf("ValuesIndex:\t%d", valuesIndex)}
		splittedStrings = append(valuesIndexString, splittedStrings...)
	}

	// Only shown manifest index if it is not -1
	if manifestIndex >= 0 {
		manifestIndexString := []string{fmt.Sprintf("DocumentIndex:\t%d", manifestIndex)}
		splittedStrings = append(manifestIndexString, splittedStrings...)
	}

	return splittedStrings
}

// diff return diff result for assertion
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

// uniform the content without invalid characters and correct line-endings
func uniformContent(content interface{}) string {
	actual := fmt.Sprintf("%v", content)
	return regex.ReplaceAllString(actual, "\n")
}

// Validate a subset, which are used for SubsetValidator and Contains (when Any option is used)
func validateSubset(actual map[string]interface{}, content map[string]interface{}) bool {
	for key := range content {
		if !reflect.DeepEqual(actual[key], content[key]) {
			return false
		}
	}

	return true
}

// Determine if the original value still is a success.
func determineSuccess(idx int, originalValue, newValue bool) bool {
	if idx == 0 {
		return newValue
	}

	return originalValue && newValue
}

func validateRequiredField(actual, field string) error {
	if actual == "" {
		return fmt.Errorf("expected field '%s' to be filled", field)
	}
	return nil
}
