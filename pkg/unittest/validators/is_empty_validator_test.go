package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

var docWithEmptyElements = `
a:
b: ""
c: 0
d: null
e: []
f: {}
`

var docWithNonEmptyElement = `
a: {a: A}
b: "b"
c: 1
d: [d]
`

func TestIsEmptyValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	for key := range manifest {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsEmptyValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docWithNonEmptyElement)

	for key := range manifest {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs:     []common.K8sManifest{manifest},
			Negative: true,
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsEmptyValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docWithNonEmptyElement)

	log.SetLevel(log.DebugLevel)

	for key, value := range manifest {
		validator := IsEmptyValidator{key}
		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})
		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"Path:	" + key,
			"Expected to be empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsEmptyValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	for key, value := range manifest {
		validator := IsEmptyValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs:     []common.K8sManifest{manifest},
			Negative: true,
		})

		marshaledValue, _ := yaml.Marshal(value)
		valueYAML := string(marshaledValue)

		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"Path:	" + key,
			"Expected NOT to be empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsEmptyValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsEmptyValidator{"a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest},
		Index: 2,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	documentIndex 2 out of range",
	}, diff)
}

func TestIsEmptyValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsEmptyValidator{"x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"a\"] from a non map type:",
		"	null",
	}, diff)
}
