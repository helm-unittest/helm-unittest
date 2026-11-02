package validators_test

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
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

func TestIsNullOrEmptyValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	for key := range manifest {
		validator := IsNullOrEmptyValidator{Path: key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsNullOrEmptyValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docWithNonEmptyElement)

	for key := range manifest {
		validator := IsNullOrEmptyValidator{Path: key}
		pass, diff := validator.Validate(&ValidateContext{
			Docs:     []common.K8sManifest{manifest},
			Negative: true,
		})

		assert.True(t, pass)
		assert.Equal(t, []string{}, diff)
	}
}

func TestIsNullOrEmptyValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docWithNonEmptyElement)

	log.SetLevel(log.DebugLevel)

	for key, value := range manifest {
		validator := IsNullOrEmptyValidator{Path: key}
		valueYAML := common.TrustedMarshalYAML(value)
		pass, diff := validator.Validate(&ValidateContext{
			Docs: []common.K8sManifest{manifest},
		})
		assert.False(t, pass)
		assert.Equal(t, []string{
			"DocumentIndex:	0",
			"ValuesIndex:	0",
			"Path:	" + key,
			"Expected to be null or empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsNullOrEmptyValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	for key, value := range manifest {
		validator := IsNullOrEmptyValidator{Path: key}
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
			"Expected NOT to be null or empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestIsNullOrEmptyValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsNullOrEmptyValidator{Path: "x.a"}
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

func TestIsNullOrEmptyValidatorWhenInvalidPathNegative(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsNullOrEmptyValidator{Path: "x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsNullOrEmptyValidatorWhenInvalidPathFailFast(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsNullOrEmptyValidator{Path: "x.a"}
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

func TestIsNullOrEmptyValidatorWhenFailFast(t *testing.T) {
	manifest := makeManifest(docWithNonEmptyElement)

	log.SetLevel(log.DebugLevel)

	for key, value := range manifest {
		validator := IsNullOrEmptyValidator{Path: key}
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
			"Expected to be null or empty, got:",
			"\t" + string(valueYAML)[:len(valueYAML)-1],
		}, diff)
	}
}

func TestFailWhenInvalidJsonPath(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsNullOrEmptyValidator{Path: "x[b]"}
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

func TestFailWhenInvalidJsonPathFailFast(t *testing.T) {
	manifest := makeManifest(docWithEmptyElements)

	validator := IsNullOrEmptyValidator{Path: "x[b]"}
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

func TestIsNullOrEmptyValidatorWhenNoManifestFail(t *testing.T) {
	validator := IsNullOrEmptyValidator{Path: "key"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\tkey",
		"Expected to be null or empty, got:",
		"\tno manifest found",
	}, diff)
}

func TestIsNullOrEmptyValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := IsNullOrEmptyValidator{Path: "key"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// The emptiness check recognises every empty form JSON can produce once parsed.
// Unparsed, the actual is the raw JSON text, which is never empty, so each case
// depends on parsing having run.
func TestIsNullOrEmptyValidatorParsedEmptyValues(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		innerPath string
	}{
		{name: "empty string", content: `{"name":""}`, innerPath: "name"},
		{name: "zero number", content: `{"replicas":0}`, innerPath: "replicas"},
		{name: "empty array", content: `{"servers":[]}`, innerPath: "servers"},
		{name: "empty object", content: `{"opts":{}}`, innerPath: "opts"},
		{name: "null value", content: `{"maybe":null}`, innerPath: "maybe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest("data:\n  config.json: |\n    " + tt.content + "\n")

			validator := IsNullOrEmptyValidator{
				Path:         `data["config.json"]`,
				ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: tt.innerPath},
			}

			pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestIsNullOrEmptyValidatorParsedNonEmptyFails(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"name":"svc"}
`)

	validator := IsNullOrEmptyValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, "svc",
		"the failure must show the decoded value, proving parsing ran")
	assert.NotContains(t, joined, `{"name"`,
		"the failure must show the decoded value, not the raw unparsed JSON text")
}

// Covers the isNotEmpty / isNotNullOrEmpty antonym path.
func TestIsNullOrEmptyValidatorParsedNonEmptyPassesNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    name: svc
`)

	validator := IsNullOrEmptyValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	// And an EMPTY value must fail the negative form, proving the emptiness
	// check really ran against the parsed value.
	emptyManifest := makeManifest(`
data:
  config.yaml: |
    name: ""
`)
	pass, _ = validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{emptyManifest},
		Negative: true,
	})
	assert.False(t, pass)
}

func TestIsNullOrEmptyValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {oops
`)

	validator := IsNullOrEmptyValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}

func TestIsNullOrEmptyValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"a":1}
`)

	validator := IsNullOrEmptyValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "missing"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, "unknown path")
	assert.Contains(t, joined, `data["config.json"]`)
	assert.Contains(t, joined, "missing")

	pass, _ = validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass, "an unmatched innerPath passes a negative assertion")
}

// innerPath may match several nodes; every one must be empty.
func TestIsNullOrEmptyValidatorParsedFanOut(t *testing.T) {
	allEmpty := makeManifest(`
data:
  config.json: |
    {"servers":[{"tag":""},{"tag":""}]}
`)

	validator := IsNullOrEmptyValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].tag"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{allEmpty}})
	assert.True(t, pass, "both tags are empty")
	assert.Equal(t, []string{}, diff)

	oneFilled := makeManifest(`
data:
  config.json: |
    {"servers":[{"tag":""},{"tag":"x"}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{oneFilled}})
	assert.False(t, pass, "one tag is non-empty, so the assertion fails")
}
