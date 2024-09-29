package valueutils_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
)

func TestGetValueOfSetPathWithSingleResults(t *testing.T) {
	a := assert.New(t)
	data := common.K8sManifest{
		"a": map[string]interface{}{
			"b":   []interface{}{"_", map[string]interface{}{"c": "yes"}},
			"d":   "no",
			"e.f": "false",
			"g":   map[string]interface{}{"h": "\"quotes\""},
			"i":   []interface{}{map[string]interface{}{"i1": "1"}, map[string]interface{}{"i2": "2"}},
			"j":   []interface{}{map[string]interface{}{"k": "1"}, map[string]interface{}{"k": "2"}},
		},
	}

	var expectionsMapping = map[string]interface{}{
		"a.b[1].c":              "yes",
		"a.b[0]":                "_",
		"a.b":                   []interface{}{"_", map[string]interface{}{"c": "yes"}},
		"a['d']":                "no",
		"a[\"e.f\"]":            "false",
		"a.g.h":                 "\"quotes\"",
		"":                      data,
		"a.i[?(@.i1 == \"1\")]": map[string]interface{}(map[string]interface{}{"i1": "1"}),
	}

	for path, expect := range expectionsMapping {
		actual, err := GetValueOfSetPath(data, path)
		a.Equal(expect, actual[0])
		a.Nil(err)
	}

    // Test jsonpath returning an array
    actual, err := GetValueOfSetPath(data, "a.j[*].k")
    a.Equal([]interface{}{"1", "2"}, actual)
    a.Nil(err)
}

func TestGetValueOfSetPathError(t *testing.T) {
	a := assert.New(t)
	data := common.K8sManifest{
		"a": map[string]interface{}{
			"b":   []interface{}{"_"},
			"c.d": "no",
		},
	}

	var expectionsMapping = map[string]string{
		"a[null]":  "invalid array index [null] before position 7: non-integer array index",
		"a.b[0[]]": "invalid array index [0[] before position 7: non-integer array index",
		"a.[c[0]]": "child name missing at position 2, following \"a.\"",
	}

	for path, expect := range expectionsMapping {
		actual, err := GetValueOfSetPath(data, path)
		a.Nil(actual)
		a.EqualError(err, expect)
	}
}

func TestBuildValueOfSetPath(t *testing.T) {
	a := assert.New(t)
	data := map[string]interface{}{"foo": "bar"}

	var expectionsMapping = map[string]interface{}{
		"a.b":    map[string]interface{}{"a": map[string]interface{}{"b": data}},
		"a[1]":   map[string]interface{}{"a": []interface{}{nil, data}},
		"a[1].b": map[string]interface{}{"a": []interface{}{nil, map[string]interface{}{"b": data}}},
	}

	for path, expected := range expectionsMapping {
		actual, err := BuildValueOfSetPath(data, path)
		a.Equal(expected, actual)
		a.Nil(err)
	}
}

func TestBuildValueSetPathError(t *testing.T) {
	a := assert.New(t)
	data := map[string]interface{}{"foo": "bar"}

	var expectionsMapping = map[string]string{
		"":   "set path is empty",
		"{":  "invalid token found {",
		"[[": "invalid escaping token [",
		"..": "unexpected end of",
	}

	for path, expect := range expectionsMapping {
		actual, err := BuildValueOfSetPath(data, path)
		a.Nil(actual)
		a.EqualError(err, expect)
	}
}

func TestMergeValues(t *testing.T) {
	a := assert.New(t)
	dest := map[string]interface{}{
		"a": map[string]interface{}{
			"b":   []interface{}{"_", map[string]interface{}{"c": "yes"}},
			"e.f": "false",
		},
	}
	src := map[string]interface{}{
		"a": map[string]interface{}{
			"b":   []interface{}{"_", map[string]interface{}{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	expected := map[string]interface{}{
		"a": map[string]interface{}{
			"b":   []interface{}{"_", map[string]interface{}{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	actual := MergeValues(dest, src)
	a.Equal(expected, actual)
}
