package helmtest_test

import (
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
)

func TestGetValueOfSetPath(t *testing.T) {
	a := assert.New(t)
	data := map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"b": []interface{}{"_", map[interface{}]interface{}{"c": "yes"}},
		},
	}

	var expectionsMapping = map[string]interface{}{
		"a.b[1].c": "yes",
		"a.b[0]":   "_",
		"a.b":      []interface{}{"_", map[interface{}]interface{}{"c": "yes"}},
	}

	for path, expect := range expectionsMapping {
		result, err := GetValueOfSetPath(data, path)
		a.Equal(result, expect)
		a.Nil(err)
	}
}
