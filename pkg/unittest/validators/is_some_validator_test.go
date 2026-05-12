package validators_test

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
)

var docWithNullElements = `
a:
b: null
`

var docWithNonNullElements = `
a: {a: A}
b: "b"
c: 1
d: [d]
e: ""
f: 0
g: []
h: {}
`

func TestIsSomeValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	for key := range manifest {
		validator := IsSomeValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsSomeValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docWithNullElements)

	for key := range manifest {
		validator := IsSomeValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs:     []common.K8sManifest{manifest},
			Negative: true,
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsSomeValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docWithNullElements)

	log.SetLevel(log.DebugLevel)

	for key, value := range manifest {
		validator := IsSomeValidator{key}
		valueYAML := common.TrustedMarshalYAML(value)
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})
		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"ValuesIndex:	0",
			"Path:	" + key,
			"Expected to be something, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsSomeValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	for key, value := range manifest {
		validator := IsSomeValidator{key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs:     []common.K8sManifest{manifest},
			Negative: true,
		})

		valueYAML := common.TrustedMarshalYAML(value)

		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"ValuesIndex:	0",
			"Path:	" + key,
			"Expected NOT to be something, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsSomeValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{"x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path x.a",
	}, diff)
}

func TestIsSomeValidatorWhenInvalidPathNegative(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{"x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSomeValidatorWhenInvalidPathFailFast(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{"x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path x.a",
	}, diff)
}

func TestIsSomeValidatorWhenFailFast(t *testing.T) {
	manifest := makeManifest(docWithNullElements)

	log.SetLevel(log.DebugLevel)

	for key, value := range manifest {
		validator := IsSomeValidator{key}
		valueYAML := common.TrustedMarshalYAML(value)
		pass, diff := validator.Validate(&ValidateContext{
			FailFast: true,
			Docs:     []common.K8sManifest{manifest, manifest},
		})
		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"ValuesIndex:	0",
			"Path:	" + key,
			"Expected to be something, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsSomeValidatorFailWhenInvalidJsonPath(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{"x[b]"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Error:",
		"\tinvalid array index [b] before position 4: non-integer array index",
		"DocumentIndex:\t1",
		"Error:",
		"\tinvalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestIsSomeValidatorFailWhenInvalidJsonPathFailFast(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{"x[b]"}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Error:",
		"\tinvalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestIsSomeValidatorWhenNoManifestFail(t *testing.T) {
	validator := IsSomeValidator{"key"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\tkey",
		"Expected to be something, got:",
		"\tno manifest found",
	}, diff)
}

func TestIsSomeValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := IsSomeValidator{"key"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
