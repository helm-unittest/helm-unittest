package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/stretchr/testify/assert"
)

func TestEqualOrGreaterValidatorOk(t *testing.T) {
	tests := []struct {
		name        string
		doc         string
		path        string
		value       interface{}
		expected    bool
		expectedErr []string
	}{
		{
			name:     "test case 1: int values",
			doc:      "spec: 4",
			path:     "spec",
			value:    5,
			expected: true,
		},
		{
			name:     "test case 2: float64 values",
			doc:      "cpu: 0.6",
			path:     "cpu",
			value:    0.75,
			expected: true,
		},
		{
			name:     "test case 3: string values",
			doc:      "cpu: 600m",
			path:     "cpu",
			value:    "600m",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest(tt.doc)

			v := EqualOrGreaterValidator{
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

func TestEqualOrGreaterValidatorFail(t *testing.T) {
	tests := []struct {
		name, doc, path string
		value           interface{}
		errorMsg        []string
	}{
		{
			name:  "test case 1: int values",
			doc:   "value: 25",
			path:  "value",
			value: 5,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"Path:\tvalue",
				"Expected to be greater then or equal to, got:",
				"\tthe expected '5' is not greater or equal to the actual '25'",
			},
		},
		{
			name:  "test case 2: float64 values",
			doc:   "cpu: 1.7",
			path:  "cpu",
			value: 1.31,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe expected '1.31' is not greater or equal to the actual '1.7'",
			},
		},
		{
			name:  "test case 3: float64 values",
			doc:   "cpu: 1.341",
			path:  "cpu",
			value: 1.338,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe expected '1.338' is not greater or equal to the actual '1.341'",
			},
		},
		{
			name:  "test case 4: string values",
			doc:   "cpu: 600m",
			path:  "cpu",
			value: "590m",
			errorMsg: []string{
				"DocumentIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe expected '590m' is not greater or equal to the actual '600m'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest(tt.doc)

			v := EqualOrGreaterValidator{
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

func TestEqualOrGreaterValidatorWhenUnkownPath(t *testing.T) {
	var actual = `
spec:
  containers:
	- name: nginx
	  image: nginx
	  resources:
		limits:
		  memory: "256Mi"
		requests:
		  cpu: 0.4
		  memory: "128Mi"
`
	manifest := makeManifest(actual)

	v := EqualOrGreaterValidator{
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

func TestEqualOrGreaterValidatorWhenTypesDoNotMatch(t *testing.T) {
	// var actual = "value: 0.3"
	// manifest := makeManifest(actual)

	tests := []struct {
		name, doc, path string
		value           interface{}
		errorMsg        []string
	}{
		{
			name:     "test case 1: compare int and string types",
			doc:      "value: 500m",
			path:     "value",
			value:    5,
			errorMsg: []string{"DocumentIndex:	0", "Error:", "	actual 'string' and expected 'int' types do not match"},
		},
		{
			name:     "test case 1: compare string and int types",
			doc:      "value: 50",
			path:     "value",
			value:    "50m",
			errorMsg: []string{"DocumentIndex:	0", "Error:", "	actual 'int' and expected 'string' types do not match"},
		},
		{
			name:     "test case 1: compare string and string(int) types",
			doc:      "value: 50",
			path:     "value",
			value:    "50",
			errorMsg: []string{"DocumentIndex:	0", "Error:", "	actual 'int' and expected 'string' types do not match"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest(tt.doc)

			v := EqualOrGreaterValidator{
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
