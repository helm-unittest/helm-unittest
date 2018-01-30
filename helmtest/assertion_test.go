package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestAssertionUnmarshaledFromYAML(t *testing.T) {
	assertionsYAML := `
- equal:
- notEqual:
- matchRegex:
- notMatchRegex:
- contains:
- notContains:
- isNull:
- isNotNull:
- isEmpty:
- isNotEmpty:
- isKind:
- isAPIVersion:
- hasDocuments:
`
	assertionsAsMap := make([]map[string]interface{}, 13)
	yaml.Unmarshal([]byte(assertionsYAML), &assertionsAsMap)
	assertions := make([]Assertion, 13)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for idx, assertion := range assertions {
		_, ok := assertionsAsMap[idx][assertion.AssertType]
		a.True(ok)
		if idx >= 10 || idx%2 == 0 {
			a.False(assertion.Not)
		} else {
			a.True(assertion.Not)
		}
	}
}

func TestAssertionUnmarshaledFromYAMLWithNotTrue(t *testing.T) {
	assertionsYAML := `
- equal:
  not: true
- notEqual:
  not: true
- matchRegex:
  not: true
- notMatchRegex:
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
`
	assertions := make([]Assertion, 13)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for idx, assertion := range assertions {
		if idx >= 10 || idx%2 == 0 {
			a.True(assertion.Not)
		} else {
			a.False(assertion.Not)
		}
	}
}

func TestReverseAssertionTheSameAsOriginalOneWithNotTrue(t *testing.T) {
	assertionsYAML := `
- equal:
	not: true
- notEqual:
- matchRegex:
	not: true
- notMatchRegex:
- contains:
	not: true
- notContains:
- isNull:
	not: true
- isNotNull:
- isEmpty:
	not: true
- isNotEmpty:
`
	assertions := make([]Assertion, 10)
	yaml.Unmarshal([]byte(assertionsYAML), &assertions)

	a := assert.New(t)
	for idx := 0; idx < len(assertions); idx += 2 {
		a.Equal(assertions[idx], assertions[idx+1])
	}
}
