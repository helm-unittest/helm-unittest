package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassNameFromPath(t *testing.T) {
	cases := map[string]string{
		"demo/templates/cm.yaml":         "demo.templates.cm_yaml",
		"demo/templates/sub/dir/cm.yaml": "demo.templates.sub.dir.cm_yaml",
		"cm.yaml":                        "cm_yaml",
	}
	for in, want := range cases {
		assert.Equal(t, want, classNameFromPath(in), in)
	}
}
