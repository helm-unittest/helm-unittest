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

// SnapshotComparer provide CompareToSnapshot utility to validator
type SnapshotComparer interface {
	CompareToSnapshot(content interface{}) *snapshot.CompareResult
}

// ValidateContext the context passed to validators
type ValidateContext struct {
	Docs     []common.K8sManifest
	Index    int
	Negative bool
	SnapshotComparer
	RenderError error
}

func (c *ValidateContext) getManifests() ([]common.K8sManifest, error) {
	manifests := make([]common.K8sManifest, 0)
	if c.Index == -1 {
		manifests = append(manifests, c.Docs...)
		return manifests, nil
	}

	if len(c.Docs) <= c.Index {
		return nil, fmt.Errorf("documentIndex %d out of range", c.Index)
	}
	manifests = append(manifests, c.Docs[c.Index])
	return manifests, nil
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
func splitInfof(format string, index int, replacements ...string) []string {
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

	if index >= 0 {
		indexedString := []string{fmt.Sprintf("DocumentIndex:\t%d", index)}
		splittedStrings = append(indexedString, splittedStrings...)
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
	// For multilines, remove spaces before newlines
	// And ensure all decoded content uses LF
	regex := regexp.MustCompile(`(?m)[ ]+\r?\n`)
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

// compareValues performs a validation of a Kubernetes manifest against an expected value.
// It compares the actual value retrieved from the manifest with the expected value,
// ensuring that they are of compatible types and that the actual value is greater, less than or equal to the expected value.
// If successful, it returns true along with a nil error slice. If unsuccessful, it returns false
// along with an error slice containing information about the validation failure.
func compareValues(expected, actual interface{}, comparisonType string) (bool, []string) {
	expStr := fmt.Sprintf("%v", expected)
	actStr := fmt.Sprintf("%v", actual)

	switch exp := expected.(type) {
	case string:
		if (comparisonType == "greater" && exp >= actual.(string)) || (comparisonType == "less" && exp <= actual.(string)) {
			return true, nil
		}
	case int:
		if (comparisonType == "greater" && exp >= actual.(int)) || (comparisonType == "less" && exp <= actual.(int)) {
			return true, nil
		}
	case float64:
		if (comparisonType == "greater" && exp >= actual.(float64)) || (comparisonType == "less" && exp <= actual.(float64)) {
			return true, nil
		}
	default:
		return false, []string{fmt.Sprintf("unsupported type '%T'", expected)}
	}

	return false, []string{fmt.Sprintf("the expected '%s' is not %s or equal to the actual '%s'", expStr, comparisonType, actStr)}
}
