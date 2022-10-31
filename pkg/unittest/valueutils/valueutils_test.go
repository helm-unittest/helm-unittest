package valueutils_test

import (
	"testing"

	. "github.com/lrills/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
)

func TestGetValueOfSetPath(t *testing.T) {
	a := assert.New(t)
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b":   []interface{}{"_", map[string]interface{}{"c": "yes"}},
			"d":   "no",
			"e.f": "false",
			"g":   map[string]interface{}{"h": "\"quotes\""},
			"i":   []interface{}{map[string]interface{}{"i1": "1"}, map[string]interface{}{"i2": "2"}},
		},
	}

	var expectionsMapping = map[string]interface{}{
		"a.b[1].c":   "yes",
		"a.b[0]":     "_",
		"a.b":        []interface{}{"_", map[string]interface{}{"c": "yes"}},
		"a['d']":     "no",
		"a[\"e.f\"]": "false",
		"a.g.h":      "\"quotes\"",
		"a.x":        nil,
		"":           data,
		//"a.i[? @.i1 == \"1\"]": []interface{}([]interface{}{map[string]interface{}{"i1": "1"}}),
	}

	for path, expect := range expectionsMapping {
		actual, err := GetValueOfSetPath(data, path)
		a.Equal(expect, actual)
		a.Nil(err)
	}
}

func TestBuildValueOfSetPath(t *testing.T) {
	a := assert.New(t)
	data := map[interface{}]interface{}{"foo": "bar"}

	var expectionsMapping = map[string]interface{}{
		"a.b":    map[interface{}]interface{}{"a": map[interface{}]interface{}{"b": data}},
		"a[1]":   map[interface{}]interface{}{"a": []interface{}{nil, data}},
		"a[1].b": map[interface{}]interface{}{"a": []interface{}{nil, map[interface{}]interface{}{"b": data}}},
	}

	for path, expected := range expectionsMapping {
		actual, err := BuildValueOfSetPath(data, path)
		a.Equal(expected, actual)
		a.Nil(err)
	}
}

func TestBuildValueSetPathError(t *testing.T) {
	a := assert.New(t)
	data := map[interface{}]interface{}{"foo": "bar"}

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
	dest := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b":   []interface{}{"_", map[interface{}]interface{}{"c": "yes"}},
			"e.f": "false",
		},
	}
	src := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b":   []interface{}{"_", map[interface{}]interface{}{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	expected := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b":   []interface{}{"_", map[interface{}]interface{}{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	actual := MergeValues(dest, src)
	a.Equal(expected, actual)
}
