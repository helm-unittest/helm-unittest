package validators_test

import (
	"strings"
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
		value       any
		expected    bool
		expectedErr []string
	}{
		{
			name:     "test case 1: int values",
			doc:      "spec: 4",
			path:     "spec",
			value:    3,
			expected: true,
		},
		{
			name:     "test case 2: float64 values",
			doc:      "cpu: 0.6",
			path:     "cpu",
			value:    0.5,
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
		value           any
		errorMsg        []string
	}{
		{
			name:  "test case 1: int values",
			doc:   "value: 25",
			path:  "value",
			value: 55,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tvalue",
				"Expected to be greater then or equal to, got:",
				"\tthe actual '25' is not greater or equal to the expected '55'",
			},
		},
		{
			name:  "test case 2: float64 values",
			doc:   "cpu: 1.7",
			path:  "cpu",
			value: 1.91,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe actual '1.7' is not greater or equal to the expected '1.91'",
			},
		},
		{
			name:  "test case 3: float64 values",
			doc:   "cpu: 1.341",
			path:  "cpu",
			value: 1.348,
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe actual '1.341' is not greater or equal to the expected '1.348'",
			},
		},
		{
			name:  "test case 4: string values",
			doc:   "cpu: 600m",
			path:  "cpu",
			value: "690m",
			errorMsg: []string{
				"DocumentIndex:\t0",
				"ValuesIndex:\t0",
				"Path:\tcpu",
				"Expected to be greater then or equal to, got:",
				"\tthe actual '600m' is not greater or equal to the expected '690m'",
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

func TestEqualOrGreaterValidatorWhenInvalidPath(t *testing.T) {
	var actual = `
spec:
  containers:
    - name: nginx
      image: nginx
`
	manifest := makeManifest(actual)

	v := EqualOrGreaterValidator{
		Path:  "spec[first]",
		Value: 1.2,
	}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [first] before position 11: non-integer array index",
	}, diff)
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

func TestEqualOrGreaterValidatorWhenUnkownPathNegative(t *testing.T) {
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

	v := EqualOrGreaterValidator{
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

func TestEqualOrGreaterValidatorWhenTypesDoNotMatch(t *testing.T) {
	tests := []struct {
		name, doc, path string
		value           any
		errorMsg        []string
	}{
		{
			name:     "test case 1: compare int and string types",
			doc:      "value: 500m",
			path:     "value",
			value:    5,
			errorMsg: []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	actual 'string' and expected 'int' types do not match"},
		},
		{
			name:     "test case 1: compare string and int types",
			doc:      "value: 50",
			path:     "value",
			value:    "50m",
			errorMsg: []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	actual 'int' and expected 'string' types do not match"},
		},
		{
			name:     "test case 1: compare string and string(int) types",
			doc:      "value: 50",
			path:     "value",
			value:    "50",
			errorMsg: []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	actual 'int' and expected 'string' types do not match"},
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

func TestEqualOrGreaterValidatorWhenTypesDoNotMatchFailFast(t *testing.T) {
	tests := []struct {
		name, doc, path string
		value           any
		errorMsg        []string
	}{
		{
			name:     "test case 1: compare int and string types",
			doc:      "value: 500m",
			path:     "value",
			value:    5,
			errorMsg: []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	actual 'string' and expected 'int' types do not match"},
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
				FailFast: true,
				Docs:     []common.K8sManifest{manifest, manifest},
			})

			assert.False(t, pass)
			assert.Equal(t, tt.errorMsg, diff)
		})
	}
}

func TestEqualOrGreaterValidatorWhenNoManifestFail(t *testing.T) {
	v := EqualOrGreaterValidator{
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
		"Expected to be greater then or equal to, got:",
		"\tno manifests found",
	}, diff)
}

func TestEqualOrGreaterValidatorWhenNoManifestNegativeOk(t *testing.T) {
	v := EqualOrGreaterValidator{
		Path:  "a.*",
		Value: 2,
	}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualOrGreaterValidatorParsedJSONNumber(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"},
		Value:        1024,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualOrGreaterValidatorParsedFloat(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"ratio":2.5}
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "ratio"},
		Value:        1.5,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualOrGreaterValidatorParsedYAMLBelowThresholdFails(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    server:
      port: 80
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "server.port"},
		Value:        1024,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), "80",
		"the failure must report the decoded number, proving parsing ran")
}

// Without normalization a parsed JSON number would arrive as float64 and this
// validator would reject the comparison outright on a type mismatch rather than
// comparing. This test pins that normalization keeps them comparable.
func TestEqualOrGreaterValidatorParsedNumberTypeIsComparable(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"port":8080}
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Value:        8080,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass, "equal values satisfy greaterOrEqual")
	assert.NotContains(t, strings.Join(diff, "\n"), "types do not match")
}

func TestEqualOrGreaterValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {nope
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Value:        1,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}

func TestEqualOrGreaterValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"a":1}
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "missing"},
		Value:        1,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, "unknown path")
	assert.Contains(t, joined, `data["config.json"]`)
	assert.Contains(t, joined, "missing")
}

// innerPath may match several nodes; every one must satisfy the comparison.
func TestEqualOrGreaterValidatorParsedFanOut(t *testing.T) {
	allAbove := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080},{"port":9090}]}
`)

	validator := EqualOrGreaterValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        1024,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{allAbove}})
	assert.True(t, pass, "both ports exceed 1024")
	assert.Equal(t, []string{}, diff)

	oneBelow := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080},{"port":80}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{oneBelow}})
	assert.False(t, pass, "one port is below the threshold, so the assertion fails")
}
