package validators

import (
	"fmt"
	"strings"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/pmezard/go-difflib/difflib"
)

type AssertInfoProvider interface {
	GetManifest(manifests []common.K8sManifest) (common.K8sManifest, error)
	IsNegative() bool
}

type Validatable interface {
	Validate(docs []common.K8sManifest, assert AssertInfoProvider) (bool, []string)
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
