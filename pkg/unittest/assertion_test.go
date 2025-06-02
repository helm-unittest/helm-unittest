package unittest_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
)

func validateSucceededTestAssertions(
	t *testing.T,
	assertionsYAML string,
	assertionCount int,
	renderedMap map[string][]common.K8sManifest,
	didPostRender bool,
) {

	assertions := make([]Assertion, assertionCount)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertions, t)

	a := assert.New(t)

	for idx, assertion := range assertions {
		cfg := AssertionConfigBuilder{
			TemplatesResult:  renderedMap,
			SnapshotComparer: fakeSnapshotComparer(true),
			RenderSucceed:    true,
			DidPostRender:    didPostRender,
		}
		assertion.WithConfig(cfg.Build())
		result := assertion.Assert(&results.AssertionResult{Index: idx})
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

func TestAssertionUnmarshalFromYAML(t *testing.T) {
	assertionsYAML := `
- equal:
- notEqual:
- greaterOrEqual:
- notGreaterOrEqual:
- lessOrEqual:
- notLessOrEqual:
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
- exists:
- notExists:
- isNullOrEmpty:
- isNotNullOrEmpty:
- isKind:
- isAPIVersion:
- isType:
- isNotType:
- hasDocuments:
- isSubset:
- isNotSubset:
- failedTemplate:
- notFailedTemplate:
- containsDocument:
- lengthEqual:
- matchSnapshot:
- matchSnapshotRaw:
`

	a := assert.New(t)
	assertionsAsMap := make([]map[string]interface{}, 33)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertionsAsMap, t)

	assertions := make([]Assertion, 33)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertions, t)

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
- greaterOrEqual:
  not: true
- notGreaterOrEqual:
  not: true
- lessOrEqual:
  not: true
- notLessOrEqual:
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
- exists:
  not: true
- notExists:
  not: true
- isNullOrEmpty:
  not: true
- isNotNullOrEmpty:
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
- isType:
  not: true
- isNotType:
  not: true
- hasDocuments:
  not: true
- isSubset:
  not: true
- failedTemplate:
  not: true
`
	a := assert.New(t)

	assertions := make([]Assertion, 29)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertions, t)

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
- greaterOrEqual:
  not: true
- notGreaterOrEqual:
- lessOrEqual:
  not: true
- notLessOrEqual:
- matchRegex:
  not: true
- notMatchRegex:
- matchRegexRaw:
  not: true
- notMatchRegexRaw:
- contains:
  not: true
- notContains:
- exists:
  not: true
- notExists:
- isNullOrEmpty:
  not: true
- isNotNullOrEmpty:
- isNull:
  not: true
- isNotNull:
- isEmpty:
  not: true
- isNotEmpty:
- isSubset:
  not: true
- isNotSubset:
- isType:
  not: true
- isNotType:
- failedTemplate:
  not: true
- notFailedTemplate:
`
	a := assert.New(t)

	assertions := make([]Assertion, 28)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertions, t)

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
x:
`
	manifest := common.TrustedUnmarshalYAML(manifestDoc)
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
  exists:
    path: x
- template: t.yaml
  notExists:
    path: g
- template: t.yaml
  isNotNullOrEmpty:
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
	validateSucceededTestAssertions(t, assertionsYAML, 15, renderedMap, false)
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
	validateSucceededTestAssertions(t, assertionsYAML, 5, renderedMap, false)
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
	common.YmlUnmarshalTestHelper(assertionYAML, &assertion, t)

	a := assert.New(t)

	cfg := AssertionConfigBuilder{
		TemplatesResult:  renderedMap,
		SnapshotComparer: fakeSnapshotComparer(true),
		RenderSucceed:    true,
	}
	assertion.WithConfig(cfg.Build())
	result := assertion.Assert(&results.AssertionResult{Index: 0})
	a.Equal(&results.AssertionResult{
		Index:      0,
		FailInfo:   []string{"Error:", "\ttemplate \"not-existed.yaml\" not exists or not selected in test suite"},
		Passed:     false,
		AssertType: "equal",
		Not:        false,
		CustomInfo: "",
	}, result)
}

func TestAssertionAssertWhenPostRendererRetainsFileSplitter(t *testing.T) {
	manifestDoc := `
kind: Fake
apiVersion: v123
a: b
c: [d]
e:
  f: g
x:
`
	manifest := common.TrustedUnmarshalYAML(manifestDoc)
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
  exists:
    path: x
- template: t.yaml
  notExists:
    path: g
- template: t.yaml
  isNotNullOrEmpty:
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
	validateSucceededTestAssertions(t, assertionsYAML, 15, renderedMap, true)
}

func TestAssertionAssertWhenPostRendererDoesNotRetainFileSplitter(t *testing.T) {
	manifestDoc := `
kind: Fake
apiVersion: v123
a: b
c: [d]
e:
  f: g
x:
`
	manifest := common.TrustedUnmarshalYAML(manifestDoc)
	renderedMap := map[string][]common.K8sManifest{
		"manifest.yaml": {manifest},
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
  exists:
    path: x
- template: t.yaml
  notExists:
    path: g
- template: t.yaml
  isNotNullOrEmpty:
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
	validateSucceededTestAssertions(t, assertionsYAML, 15, renderedMap, true)
}

func TestAssertionAssertWhenTemplateNotSpecifiedAndNoDefault(t *testing.T) {
	manifest := common.K8sManifest{}
	renderedMap := map[string][]common.K8sManifest{
		"existed.yaml": {manifest},
	}
	assertionYAML := "equal:"
	assertion := new(Assertion)
	common.YmlUnmarshalTestHelper(assertionYAML, &assertion, t)

	a := assert.New(t)
	cfg := AssertionConfigBuilder{
		TemplatesResult:  renderedMap,
		SnapshotComparer: fakeSnapshotComparer(true),
		RenderSucceed:    true,
	}
	assertion.WithConfig(cfg.Build())
	result := assertion.Assert(&results.AssertionResult{Index: 0})
	a.Equal(&results.AssertionResult{
		Index:      0,
		FailInfo:   []string{"Error:", "\tassertion.template must be given if testsuite.templates is empty"},
		Passed:     false,
		AssertType: "equal",
		Not:        false,
		CustomInfo: "",
	}, result)
}

func TestAssertionAssertWhenDocumentIndexIsOutOfRange(t *testing.T) {
	manifest := common.K8sManifest{}
	renderedMap := map[string][]common.K8sManifest{
		"template.yaml": {manifest},
	}
	assertionYAML := `
template: template.yaml
documentIndex: 1
equal:
`
	assertion := new(Assertion)
	common.YmlUnmarshalTestHelper(assertionYAML, &assertion, t)

	a := assert.New(t)
	cfg := AssertionConfigBuilder{
		TemplatesResult:  renderedMap,
		SnapshotComparer: fakeSnapshotComparer(true),
		RenderSucceed:    true,
	}
	assertion.WithConfig(cfg.Build())
	result := assertion.Assert(&results.AssertionResult{Index: 0})
	a.Equal(&results.AssertionResult{
		Index:      0,
		FailInfo:   []string{"Error:", "document index 1 is out of range"},
		Passed:     false,
		AssertType: "equal",
		Not:        false,
		CustomInfo: "",
	}, result)
}

func TestAssertionWithSkippedDocument(t *testing.T) {
	manifestDoc := `
kind: Fake
apiVersion: v123
a: b
e:
  f: g
x:
`
	manifest := common.TrustedUnmarshalYAML(manifestDoc)
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
`
	assertions := make([]Assertion, 2)
	common.YmlUnmarshalTestHelper(assertionsYAML, &assertions, t)

	ds := &valueutils.DocumentSelector{
		Path:               "kind",
		Value:              "NotExist",
		SkipEmptyTemplates: true,
	}
	cfg := AssertionConfigBuilder{
		TemplatesResult:     renderedMap,
		SnapshotComparer:    fakeSnapshotComparer(true),
		RenderSucceed:       true,
		IsSkipEmptyTemplate: ds.SkipEmptyTemplates,
	}
	for _, assertion := range assertions {
		assertion.DocumentSelector = ds
		assertion.WithConfig(cfg.Build())

		result := assertion.Assert(&results.AssertionResult{Index: 0})
		assert.True(t, result.Passed)
		assert.True(t, result.Skipped)
	}
}
