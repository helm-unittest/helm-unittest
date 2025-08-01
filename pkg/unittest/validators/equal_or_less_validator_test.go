package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/stretchr/testify/assert"
)

func TestEqualOrLessValidatorOk(t *testing.T) {
	tests := []struct {
		name        string
		doc         string
		path        string
		value       any
		expected    bool
		expectedErr []string
	}{
		{
			name:     "Test case 1: int ok",
			doc:      "spec: 4",
			path:     "spec",
			value:    5,
			expected: true,
		},
		{
			name:     "Test case 2: float64 ok",
			doc:      "cpu: 0.6",
			path:     "cpu",
			value:    0.7,
			expected: true,
		},
		{
			name:     "Test case 3: string ok",
			doc:      "cpu: 600m",
			path:     "cpu",
			value:    "680m",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest(tt.doc)

			v := EqualOrLessValidator{
				Path:  tt.path,
				Value: tt.value,
			}
			pass, diff := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestEqualOrLessValidatorFail(t *testing.T) {
	tests := []struct {
		name, doc, path string
		value           any
		errorMsg        []string
	}{
		{
			name:  "Test case 1: int fail",
			doc:   "value: 6",
			path:  "value",
			value: 5,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tvalue",
				"Expected to be less then or equal to, got:",
				"\tthe actual '6' is not less or equal to the expected '5'",
			},
		},
		{
			name:  "Test case 2: float64 fail",
			doc:   "cpu: 1.7",
			path:  "cpu",
			value: 1.69,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tcpu",
				"Expected to be less then or equal to, got:",
				"\tthe actual '1.7' is not less or equal to the expected '1.69'",
			},
		},
		{
			name:  "Test case 3: float64 fail",
			doc:   "cpu: 1.341",
			path:  "cpu",
			value: 1.34,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tcpu",
				"Expected to be less then or equal to, got:",
				"\tthe actual '1.341' is not less or equal to the expected '1.34'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest(tt.doc)

			v := EqualOrLessValidator{
				Path:  tt.path,
				Value: tt.value,
			}
			pass, diff := v.Validate(&ValidateContext{
				Docs: []common.K8sManifest{manifest},
			})

			assert.False(t, pass)
			assert.Equal(t, tt.errorMsg, diff)
		})
	}
}

func TestEqualOrLessValidatorWhenUnkownPath(t *testing.T) {
	var actual = `
spec:
  containers:
    - name: nginx
      image: nginx
      resources:
        limits:
          memory: "256Mi"
        requests:
          memory: "128Mi"
`
	manifest := makeManifest(actual)

	v := EqualOrLessValidator{
		Path:  "spec.containers[0].resources.requests.cpu",
		Value: 1.2,
	}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path 'spec.containers[0].resources.requests.cpu'",
	}, diff)
}

func TestEqualOrLessValidatorWhenUnkownPathNegative(t *testing.T) {
	var actual = `
spec:
  containers:
    - name: nginx
      image: nginx
      resources:
        limits:
          memory: "256Mi"
        requests:
          memory: "128Mi"
`
	manifest := makeManifest(actual)

	v := EqualOrLessValidator{
		Path:  "spec.containers[0].resources.requests.cpu",
		Value: 1.2,
	}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualOrLessValidatorWhenTypesDoNotMatch(t *testing.T) {
	var actual = "value: 0.3"
	manifest := makeManifest(actual)

	v := EqualOrLessValidator{
		Path:  "value",
		Value: 1,
	}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Error:",
		"	actual 'float64' and expected 'int' types do not match",
	}, diff)
}

func TestEqualOrLessValidatorWhenFailFast(t *testing.T) {
	var actual = `
a:
  b: 1
  c: 1
`
	manifest := makeManifest(actual)

	v := EqualOrLessValidator{
		Path:  "a.*",
		Value: 0,
	}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"ValuesIndex:\t0",
		"Path:\ta.*",
		"Expected to be less then or equal to, got:",
		"\tthe actual '1' is not less or equal to the expected '0'",
	}, diff)
}

func TestEqualOrLessValidatorWhenNoManifestFail(t *testing.T) {
	v := EqualOrLessValidator{
		Path:  "a.*",
		Value: 2,
	}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\ta.*",
		"Expected to be less then or equal to, got:",
		"\tno manifests found",
	}, diff)
}

func TestEqualOrLessValidatorWhenNoManifestNegativeOk(t *testing.T) {
	v := EqualOrLessValidator{
		Path:  "a.*",
		Value: 0,
	}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
