package validators

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/pmezard/go-difflib/difflib"
)

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

// uniform the content with correct line-endings
func uniformContent(content interface{}) string {
	// All decoded content uses LF
	actual := fmt.Sprintf("%v", content)
	return strings.ReplaceAll(actual, "\r\n", "\n")
}

const errorFormat = `
Error:
%s
`

// Validate a subset, which are used for SubsetValidator and Contains (when Any option is used)
func validateSubset(actual map[interface{}]interface{}, content interface{}) bool {
	found := false
	for key, value := range actual {
		ele := map[interface{}]interface{}{key: value}
		if reflect.DeepEqual(ele, content) {
			found = true
		}
	}
	return found
}

// Determine if the original value still is a success.
func determineSuccess(idx int, originalValue, newValue bool) bool {
	if idx == 0 {
		return true
	}

	return originalValue && newValue
}
