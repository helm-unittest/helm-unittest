package valueutils_test

import (
	"fmt"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"

	v3util "helm.sh/helm/v3/pkg/chartutil"
)

func TestGetValueOfSetPathWithSingleResults(t *testing.T) {
	a := assert.New(t)
	data := common.K8sManifest{
		"a": map[string]any{
			"b":   []any{"_", map[string]any{"c": "yes"}},
			"d":   "no",
			"e.f": "false",
			"g":   map[string]any{"h": "\"quotes\""},
			"i":   []any{map[string]any{"i1": "1"}, map[string]any{"i2": "2"}},
		},
	}

	var expectionsMapping = map[string]any{
		"a.b[1].c":              "yes",
		"a.b[0]":                "_",
		"a.b":                   []any{"_", map[string]any{"c": "yes"}},
		"a['d']":                "no",
		"a[\"e.f\"]":            "false",
		"a.g.h":                 "\"quotes\"",
		"":                      data,
		"a.i[?(@.i1 == \"1\")]": map[string]any(map[string]any{"i1": "1"}),
	}

	for path, expect := range expectionsMapping {
		actual, err := GetValueOfSetPath(data, path)
		a.Equal(expect, actual[0])
		a.Nil(err)
	}
}

func TestGetValueOfSetPathError(t *testing.T) {
	a := assert.New(t)
	data := common.K8sManifest{
		"a": map[string]any{
			"b":   []any{"_"},
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
	data := map[string]any{"foo": "bar"}

	var expectionsMapping = map[string]any{
		"a.b":    map[string]any{"a": map[string]any{"b": data}},
		"a[1]":   map[string]any{"a": []any{nil, data}},
		"a[1].b": map[string]any{"a": []any{nil, map[string]any{"b": data}}},
	}

	for path, expected := range expectionsMapping {
		actual, err := BuildValueOfSetPath(data, path)
		a.Equal(expected, actual)
		a.Nil(err)
	}
}

func TestBuildValueOfSetPath_V1(t *testing.T) {

	t.Run("path is empty", func(t *testing.T) {
		_, err := BuildValueOfSetPath(nil, "")
		assert.Error(t, err)
		assert.EqualError(t, err, "set path is empty")
	})

	t.Run("value is empty", func(t *testing.T) {
		actual, err := BuildValueOfSetPath(nil, "some.path")
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{
			"some": map[string]any{
				"path": nil,
			},
		}, actual)
	})

	t.Run("some path", func(t *testing.T) {
		expected := map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"a": 1,
					"b": map[string]any{
						"c": 2,
					}}}}
		val := map[string]any{
			"a": 1,
			"b": map[string]any{
				"c": 2,
			},
		}
		path := "b.c"
		result, err := BuildValueOfSetPath(val, path)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("path is not in data", func(t *testing.T) {
		expected := map[string]any{
			"some": map[string]any{
				"path": map[string]any{
					"ingress": map[string]any{
						"hosts[1]": "example.local",
					},
				},
			},
		}
		var data = map[string]any{
			"ingress": map[string]any{"hosts[1]": "example.local"},
		}
		actual, err := BuildValueOfSetPath(data, "some.path")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("path is in values", func(t *testing.T) {
		expected := map[string]any{
			"hosts": map[string]any{
				"ingress": map[string]any{
					"hosts[1]": "example.local",
				},
			},
		}
		var data = map[string]any{
			"ingress": map[string]any{"hosts[1]": "example.local"},
		}
		actual, err := BuildValueOfSetPath(data, "hosts")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("property testing", func(t *testing.T) {
		data := map[string]any{"foo": "bar"}
		cases := []struct {
			input  map[string]any
			path   string
			exp    map[string]any
			expStr string
		}{
			{
				path:   "a.b",
				input:  map[string]any{"a": map[string]any{"b": data}},
				exp:    map[string]any{"a": map[string]any{"b": map[string]any{"a": map[string]any{"b": data}}}},
				expStr: "map[a:map[b:map[a:map[b:map[foo:bar]]]]]",
			},
			{
				path:   "a[1]",
				input:  map[string]any{"a": []any{nil, data}},
				exp:    map[string]any{"a": []any{any(nil), map[string]any{"a": []any{any(nil), data}}}},
				expStr: "map[a:[<nil> map[a:[<nil> map[foo:bar]]]]]",
			},
			{
				path:   "a[1].b",
				input:  map[string]any{"a": []any{nil, map[string]any{"b": data}}},
				exp:    map[string]any{"a": []any{any(nil), map[string]any{"b": map[string]any{"a": []any{any(nil), map[string]any{"b": data}}}}}},
				expStr: "map[a:[<nil> map[b:map[a:[<nil> map[b:map[foo:bar]]]]]]]",
			},
		}
		for _, test := range cases {
			t.Run(fmt.Sprintf("path %s and values '%v", test.path, test.input), func(t *testing.T) {
				actual, err := BuildValueOfSetPath(test.input, test.path)
				assert.NoError(t, err)
				assert.Equal(t, test.exp, actual)
				assert.Equal(t, test.expStr, fmt.Sprintf("%v", actual))
			})
		}
	})
}

func TestBuildValueSetPathError(t *testing.T) {
	a := assert.New(t)
	data := map[string]any{"foo": "bar"}

	var expectionsMapping = map[string]string{
		"":         "set path is empty",
		"{":        "invalid token found {",
		"[[":       "invalid escaping token [",
		"..":       "unexpected end of",
		"foo[1.1]": "missing index value",
		"foo[]":    "strconv.Atoi: parsing \"\": invalid syntax",
		"foo]":     "invalid key ]",
	}

	for path, expect := range expectionsMapping {
		actual, err := BuildValueOfSetPath(data, path)
		a.Nil(actual)
		a.EqualError(err, expect)
	}
}

func TestMergeValues(t *testing.T) {
	a := assert.New(t)
	src := map[string]any{
		"a": map[string]any{
			"b":   []any{"_", map[string]any{"c": "yes"}},
			"e.f": "false",
		},
	}
	dest := map[string]any{
		"a": map[string]any{
			"b":   []any{"_", map[string]any{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	expected := map[string]any{
		"a": map[string]any{
			"b":   []any{"_", map[string]any{"c": "no"}, "a"},
			"d":   "no",
			"e.f": "yes",
		},
	}
	actual := v3util.MergeTables(dest, src)
	a.Equal(expected, actual)
}

func TestMergeValues_Cases(t *testing.T) {

	t.Run("SimpleMerge", func(t *testing.T) {
		src := map[string]any{"a": 1, "b": 2}
		dest := map[string]any{"c": 3, "d": 4}
		expected := map[string]any{"a": 1, "b": 2, "c": 3, "d": 4}
		result := v3util.MergeTables(dest, src)
		assert.Equal(t, expected, result)
	})

	t.Run("OverwriteExistingValue", func(t *testing.T) {
		src := map[string]any{"a": 1}
		dest := map[string]any{"a": 2}
		expected := map[string]any{"a": 2}
		result := v3util.MergeTables(dest, src)
		assert.Equal(t, expected, result)
	})

	t.Run("MergeNestedMaps", func(t *testing.T) {
		src := map[string]any{"a": map[string]any{"b": 1}}
		dest := map[string]any{"a": map[string]any{"c": 2}}
		expected := map[string]any{"a": map[string]any{"b": 1, "c": 2}}
		result := v3util.MergeTables(dest, src)
		assert.Equal(t, expected, result)
	})

	t.Run("OverwriteNestedMap", func(t *testing.T) {
		src := map[string]any{"a": map[string]any{"b": 1}}
		dest := map[string]any{"a": 2}
		expected := map[string]any{"a": 2}
		result := v3util.MergeTables(dest, src)
		assert.Equal(t, expected, result)
	})

	t.Run("MergeComplexMaps", func(t *testing.T) {
		src := map[string]any{
			"a": 1,
			"b": map[string]any{
				"c": 2,
			},
		}
		dest := map[string]any{
			"a": 3,
			"b": map[string]any{
				"d": 4,
			},
			"e": 5,
		}
		expected := map[string]any{
			"a": 3,
			"b": map[string]any{
				"c": 2,
				"d": 4,
			},
			"e": 5,
		}
		result := v3util.MergeTables(dest, src)
		assert.Equal(t, expected, result)
	})
}

func TestMergeValues_YamlValues(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		yamlSrc := `
a:
  b:
   foo: bar
`
		yamlDst := `
a:
  hosts[0]: abrakadabra
`
		expected := `
a:
  b:
   foo: bar
  hosts[0]: abrakadabra
`
		var dataSrc map[string]any
		common.YmlUnmarshalTestHelper(yamlSrc, &dataSrc, t)
		var dataDst map[string]any
		common.YmlUnmarshalTestHelper(yamlDst, &dataDst, t)

		output := v3util.MergeTables(dataDst, dataSrc)
		out, _ := common.YmlMarshall(&output)
		assert.YAMLEq(t, expected, out)
	})

	t.Run("second", func(t *testing.T) {
		yamlSrc := `
a:
  b:
   hosts:
   - foo
   - bar
`
		yamlDst := `
a:
  b:
   hosts[0]: abrakadabra
`
		expected := `
a:
  b:
   hosts:
   - foo
   - bar
   hosts[0]: abrakadabra
`
		var dataSrc map[string]any
		common.YmlUnmarshalTestHelper(yamlSrc, &dataSrc, t)
		var dataDst map[string]any
		common.YmlUnmarshalTestHelper(yamlDst, &dataDst, t)

		output := v3util.MergeTables(dataDst, dataSrc)
		out := common.YmlMarshallTestHelper(&output, t)
		assert.YAMLEq(t, expected, out)
	})
}

func TestGetValueOfSetPath(t *testing.T) {
	t.Run("invalid-path", func(t *testing.T) {
		yml := `
kind: Deployment
metadata:
  name: my-deployment
`
		var dataDst map[string]any
		common.YmlUnmarshalTestHelper(yml, &dataDst, t)

		actual, err := GetValueOfSetPath(dataDst, "invalid.path")
		assert.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("valid-path", func(t *testing.T) {
		yml := `
kind: Deployment
metadata:
  name: my-deployment
`
		var dataDst map[string]any
		common.YmlUnmarshalTestHelper(yml, &dataDst, t)

		actual, err := GetValueOfSetPath(dataDst, "metadata.name")
		assert.NoError(t, err)
		assert.Equal(t, []any{"my-deployment"}, actual)
	})
}

func TestBuildValueOfSetPath_EmptyPath(t *testing.T) {
	_, err := BuildValueOfSetPath(nil, "")
	assert.Error(t, err)
	assert.EqualError(t, err, "set path is empty")
}

func TestBuildValueOfSetPath_ValidPath(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	expected := map[string]any{"a": map[string]any{"b": data}}
	actual, err := BuildValueOfSetPath(data, "a.b")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestBuildValueOfSetPath_InvalidToken(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	_, err := BuildValueOfSetPath(data, "{")
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid token found {")
}

func TestBuildValueOfSetPath_UnexpectedEnd(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	_, err := BuildValueOfSetPath(data, "a[")
	assert.Error(t, err)
	assert.EqualError(t, err, "unexpected end of")
}

func TestBuildValueOfSetPath_NestedPath(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	expected := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": data,
			},
		},
	}
	actual, err := BuildValueOfSetPath(data, "a.b.c")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

// merge values
func TestMergeValues_EmptySource(t *testing.T) {
	src := map[string]any{"a": 1}
	dest := map[string]any{}
	expected := map[string]any{"a": 1}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_EmptyDestination(t *testing.T) {
	src := map[string]any{}
	dest := map[string]any{"a": 1}
	expected := map[string]any{"a": 1}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_NilDestination(t *testing.T) {
	src := map[string]any{"a": 1}
	dest := map[string]any{}
	expected := map[string]any{"a": 1}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_NilSource(t *testing.T) {
	src := map[string]any{}
	dest := map[string]any{"a": 1}
	expected := map[string]any{"a": 1}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_OverwriteWithNonMap(t *testing.T) {
	src := map[string]any{"a": map[string]any{"b": 1}}
	dest := map[string]any{"a": 2}
	expected := map[string]any{"a": 2}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_DeepMerge(t *testing.T) {
	src := map[string]any{"a": map[string]any{"b": 1}}
	dest := map[string]any{"a": map[string]any{"c": 2}}
	expected := map[string]any{"a": map[string]any{"b": 1, "c": 2}}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_ComplexMerge(t *testing.T) {
	src := map[string]any{
		"a": 1,
		"b": map[string]any{
			"c": 2,
		},
	}
	dest := map[string]any{
		"a": 3,
		"b": map[string]any{
			"d": 4,
		},
		"e": 5,
	}
	expected := map[string]any{
		"a": 3,
		"b": map[string]any{
			"c": 2,
			"d": 4,
		},
		"e": 5,
	}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

// new
func TestMergeValues_KeyExistsButNotMap(t *testing.T) {
	src := map[string]any{
		"a": 1,
	}
	dest := map[string]any{
		"a": map[string]any{"b": 2},
	}
	expected := map[string]any{
		"a": map[string]any{"b": 2},
	}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_KeyExistsAndIsMap(t *testing.T) {
	src := map[string]any{
		"a": map[string]any{"b": 1},
	}
	dest := map[string]any{
		"a": map[string]any{"c": 2},
	}
	expected := map[string]any{
		"a": map[string]any{"b": 1, "c": 2},
	}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_EmptySourceAndDestination(t *testing.T) {
	src := map[string]any{}
	dest := map[string]any{}
	expected := map[string]any{}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_DestinationWithNilValue(t *testing.T) {
	src := map[string]any{"a": 1}
	dest := map[string]any{"b": nil}
	expected := map[string]any{"a": 1, "b": nil}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_SourceWithNilValue(t *testing.T) {
	src := map[string]any{"a": nil}
	dest := map[string]any{"b": 2}
	expected := map[string]any{"a": nil, "b": 2}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_NestedMapWithEmptyMap(t *testing.T) {
	src := map[string]any{"a": map[string]any{"b": 1}}
	dest := map[string]any{"a": map[string]any{}}
	expected := map[string]any{"a": map[string]any{"b": 1}}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_EmptyNestedMap(t *testing.T) {
	src := map[string]any{"a": map[string]any{}}
	dest := map[string]any{"a": map[string]any{"b": 1}}
	expected := map[string]any{"a": map[string]any{"b": 1}}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_OverwriteWithEmptyMap(t *testing.T) {
	src := map[string]any{"a": map[string]any{"b": 1}}
	dest := map[string]any{"a": map[string]any{"b": nil}}
	expected := map[string]any{"a": map[string]any{"b": nil}}
	// expected := map[string]interface{}{"a": map[string]interface{}{}}
	// this is the expected result when v3util.CoalesceTables is used
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMergeValues_OverwriteWithNil(t *testing.T) {
	src := map[string]any{"a": map[string]any{"b": 1}}
	dest := map[string]any{"a": nil}
	expected := map[string]any{"a": nil}
	actual := v3util.MergeTables(dest, src)
	assert.Equal(t, expected, actual)
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		input    string
		pattern  string
		expected bool
		hasError bool
	}{
		{"example123", `^[a-z]+[0-9]+$`, true, false},
		{"example", `^[a-z]+[0-9]+$`, false, false},
		{"123example", `^[a-z]+[0-9]+$`, false, false},
		{"example123", `[a-z]+`, true, false},
		{"example123", `\d+`, true, false},
		{"example123", `(`, false, true}, // Invalid regex pattern
	}

	for _, test := range tests {
		t.Run(test.input+"_"+test.pattern, func(t *testing.T) {
			result, err := MatchesPattern(test.input, test.pattern)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
