package unittest_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func validateSucceededTestAssertions(
	t *testing.T,
	assertionsYAML string,
	assertionCount int,
	renderedMap map[string][]common.K8sManifest) {

	assertions := make([]Assertion, assertionCount)
	err := yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	a.Nil(err)

	for idx, assertion := range assertions {
		result := assertion.Assert(renderedMap, fakeSnapshotComparer(true), true, nil, &results.AssertionResult{Index: idx})
		a.Equal(&results.AssertionResult{
			Index:      idx,
			FailInfo:   []string{},
			Passed:     true,
			AssertType: assertion.AssertType,
			Not:        false,
			CustomInfo: "",
		}, result)
	}

}

func TestAssertionUnmarshaledFromYAML(t *testing.T) {
	assertionsYAML := `
- equal:
- notEqual:
- equalRaw:
- notEqualRaw:
- matchRegex:
- notMatchRegex:
- matchRegexRaw:
- notMatchRegexRaw:
- contains:
- notContains:
- isNull:
- isNotNull:
- isEmpty:
- isNotEmpty:
- isKind:
- isAPIVersion:
- hasDocuments:
- isSubset:
- isNotSubset:
- failedTemplate:
- notFailedTemplate:
- containsDocument:
- lengthEqual:
`

	assertionsAsMap := make([]map[string]interface{}, 23)
	yaml.Unmarshal([]byte(assertionsYAML), &assertionsAsMap)
	assertions := make([]Assertion, 23)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for idx, assertion := range assertions {
		_, ok := assertionsAsMap[idx][assertion.AssertType]
		a.True(ok)
		a.False(assertion.Not)
	}
}

func TestAssertionUnmarshaledFromYAMLWithNotTrue(t *testing.T) {
	assertionsYAML := `
- equal:
  not: true
- notEqual:
  not: true
- equalRaw:
  not: true
- notEqualRaw:
  not: true
- matchRegex:
  not: true
- notMatchRegex:
  not: true
- matchRegexRaw:
  not: true
- notMatchRegexRaw:
  not: true
- contains:
  not: true
- notContains:
  not: true
- isNull:
  not: true
- isNotNull:
  not: true
- isEmpty:
  not: true
- isNotEmpty:
  not: true
- isKind:
  not: true
- isAPIVersion:
  not: true
- hasDocuments:
  not: true
- isSubset:
  not: true
- failedTemplate:
  not: true
`
	assertions := make([]Assertion, 19)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for _, assertion := range assertions {
		a.True(assertion.Not)
	}
}

func TestReverseAssertionTheSameAsOriginalOneWithNotTrue(t *testing.T) {
	assertionsYAML := `
- equal:
  not: true
- notEqual:
- equalRaw:
  not: true
- notEqualRaw:
- matchRegex:
  not: true
- notMatchRegex:
- matchRegexRaw:
  not: true
- notMatchRegexRaw:
- contains:
  not: true
- notContains:
- isNull:
  not: true
- isNotNull:
- isEmpty:
  not: true
- isNotEmpty:
- isSubset:
  not: true
- isNotSubset:
- failedTemplate:
  not: true
- notFailedTemplate:
`
	assertions := make([]Assertion, 17)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for idx := 0; idx < len(assertions); idx += 2 {
		a.Equal(assertions[idx].Not, !assertions[idx+1].Not)
	}
}

type fakeSnapshotComparer bool

func (c fakeSnapshotComparer) CompareToSnapshot(content interface{}) *snapshot.CompareResult {
	return &snapshot.CompareResult{
		Passed: bool(c),
	}
}

func TestAssertionAssertWhenOk(t *testing.T) {
	manifestDoc := `
kind: Fake
apiVersion: v123
a: b
c: [d]
e:
  f: g
`
	manifest := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifestDoc), &manifest)
	renderedMap := map[string][]common.K8sManifest{
		"t.yaml": {manifest},
	}

	assertionsYAML := `
- template: t.yaml
  equal:
    path:  a
    value: b
- template: t.yaml
  notEqual:
    path:  a
    value: c
- template: t.yaml
  matchRegex:
    path:    a
    pattern: b
- template: t.yaml
  notMatchRegex:
    path:    a
    pattern: c
- template: t.yaml
  contains:
    path:    c
    content: d
- template: t.yaml
  notContains:
    path:    c
    content: e
- template: t.yaml
  isNull:
    path: x
- template: t.yaml
  isNotNull:
    path: a
- template: t.yaml
  isEmpty:
    path: z
- template: t.yaml
  isNotEmpty:
    path: c
- template: t.yaml
  isKind:
    of: Fake
- template: t.yaml
  isAPIVersion:
    of: v123
- template: t.yaml
  hasDocuments:
    count: 1
- template: t.yaml
  matchSnapshot: {}
- template: t.yaml
  isSubset:
    path: e
    content: 
      f: g
- template: t.yaml
  lengthEqual:
    path: c
    count: 1
`
	validateSucceededTestAssertions(t, assertionsYAML, 15, renderedMap)
}

func TestAssertionRawAssertWhenOk(t *testing.T) {
	manifest := common.K8sManifest{common.RAW: "NOTES.txt"}
	renderedMap := map[string][]common.K8sManifest{
		"t.yaml": {manifest},
	}

	assertionsYAML := `
- template: t.yaml
  equalRaw:
    value: NOTES.txt
- template: t.yaml
  notEqualRaw:
    value: UNNOTES.txt
- template: t.yaml
  matchRegexRaw:
    pattern: NOTES.txt
- template: t.yaml
  notMatchRegexRaw:
    pattern: UNNOTES.txt
- template: t.yaml
  hasDocuments:
    count: 1
- template: t.yaml
  matchSnapshot: {}
`
	validateSucceededTestAssertions(t, assertionsYAML, 5, renderedMap)
}

func TestAssertionAssertWhenTemplateNotExisted(t *testing.T) {
	manifest := common.K8sManifest{}
	renderedMap := map[string][]common.K8sManifest{
		"existed.yaml": {manifest},
	}
	assertionYAML := `
template: not-existed.yaml
equal:
`
	assertion := new(Assertion)
	err := yaml.Unmarshal([]byte(assertionYAML), &assertion)

	a := assert.New(t)
	a.Nil(err)

	result := assertion.Assert(renderedMap, fakeSnapshotComparer(true), true, nil, &results.AssertionResult{Index: 0})
	a.Equal(&results.AssertionResult{
		Index:      0,
		FailInfo:   []string{"Error:", "\ttemplate \"not-existed.yaml\" not exists or not selected in test suite"},
		Passed:     false,
		AssertType: "equal",
		Not:        false,
		CustomInfo: "",
	}, result)
}

func TestAssertionAssertWhenTemplateNotSpecifiedAndNoDefault(t *testing.T) {
	manifest := common.K8sManifest{}
	renderedMap := map[string][]common.K8sManifest{
		"existed.yaml": {manifest},
	}
	assertionYAML := "equal:"
	assertion := new(Assertion)
	yaml.Unmarshal([]byte(assertionYAML), &assertion)

	a := assert.New(t)
	result := assertion.Assert(renderedMap, fakeSnapshotComparer(true), true, nil, &results.AssertionResult{Index: 0})
	a.Equal(&results.AssertionResult{
		Index:      0,
		FailInfo:   []string{"Error:", "\tassertion.template must be given if testsuite.templates is empty"},
		Passed:     false,
		AssertType: "equal",
		Not:        false,
		CustomInfo: "",
	}, result)
}
