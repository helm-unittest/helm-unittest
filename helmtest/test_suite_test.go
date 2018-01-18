package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
)

func TestParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	s, err := ParseTestSuiteFile("../__fixtures__/basic/tests/list_all_field.yaml")

	a.Nil(err)
	a.Equal(s.Name, "test suite name")
	a.Equal(s.Files, []string{"a.yaml", "b.yaml"})
	a.Equal(s.Tests[0].Name, "should ...")
	a.Equal(s.Tests[0].Values, []string{"values.yaml"})
	a.Equal(s.Tests[0].Set, map[string]interface{}{
		"a.b.c": "ABC",
		"x.y.z": "XYZ",
	})
	// a.Equal(s.Tests[0].Assertions, []Assertion{
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "equal"},
	// 	Assertion{DocumentIndex: 0, Not: true, AssertType: "notEqual"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "matchRegex"},
	// 	Assertion{DocumentIndex: 0, Not: true, AssertType: "notMatchRegex"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "contains"},
	// 	Assertion{DocumentIndex: 0, Not: true, AssertType: "notContains"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "isNull"},
	// 	Assertion{DocumentIndex: 0, Not: true, AssertType: "isNotNull"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "isEmpty"},
	// 	Assertion{DocumentIndex: 0, Not: true, AssertType: "isNotEmpty"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "isKind"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "isAPIVersion"},
	// 	Assertion{DocumentIndex: 0, Not: false, AssertType: "hasDocuments"},
	// })
}
