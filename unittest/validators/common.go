package validators

import (
	"fmt"
	"strings"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
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

const errorFormat = `
Error:
%s
`
