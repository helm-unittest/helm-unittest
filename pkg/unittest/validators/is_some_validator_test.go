package validators_test

import (
	"strings"
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
		validator := IsSomeValidator{Path: key}
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
		validator := IsSomeValidator{Path: key}
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
		validator := IsSomeValidator{Path: key}
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
		validator := IsSomeValidator{Path: key}
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

	validator := IsSomeValidator{Path: "x.a"}
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

	validator := IsSomeValidator{Path: "x.a"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSomeValidatorWhenInvalidPathFailFast(t *testing.T) {
	manifest := makeManifest(docWithNonNullElements)

	validator := IsSomeValidator{Path: "x.a"}
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
		validator := IsSomeValidator{Path: key}
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

	validator := IsSomeValidator{Path: "x[b]"}
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

	validator := IsSomeValidator{Path: "x[b]"}
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
	validator := IsSomeValidator{Path: "key"}
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
	validator := IsSomeValidator{Path: "key"}
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
func TestIsSomeValidatorParsedEmptyValues(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		innerPath string
	}{
		{name: "empty string", content: `{"name":""}`, innerPath: "name"},
		{name: "zero number", content: `{"replicas":0}`, innerPath: "replicas"},
		{name: "empty array", content: `{"servers":[]}`, innerPath: "servers"},
		{name: "empty object", content: `{"opts":{}}`, innerPath: "opts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest("data:\n  config.json: |\n    " + tt.content + "\n")

			validator := IsSomeValidator{
				Path:         `data["config.json"]`,
				ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: tt.innerPath},
			}

			pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestIsSomeValidatorParsedNonEmptyFails(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"name":"svc"}
`)

	validator := IsSomeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// Covers the isSome antonym path.
func TestIsSomeValidatorParsedNonEmptyPassesNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    name: svc
`)

	validator := IsSomeValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	" + validator.Path,
		"Expected NOT to be something, got:",
		"\tsvc",
	}, diff)

	// And an NULL value must fail the negative form, proving the emptiness
	// check really ran against the parsed value.
	emptyManifest := makeManifest(`
data:
  config.yaml: |
    name: null
`)
	pass, _ = validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{emptyManifest},
		Negative: true,
	})
	assert.True(t, pass)
}

func TestIsSomeValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {oops
`)

	validator := IsSomeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "name"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}

func TestIsSomeValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"a":1}
`)

	validator := IsSomeValidator{
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

// innerPath may match several nodes; every one must have some value.
func TestIsSomeValidatorParsedFanOut(t *testing.T) {
	allEmpty := makeManifest(`
data:
  config.json: |
    {"servers":[{"tag":"value1"},{"tag":"value2"}]}
`)

	validator := IsSomeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].tag"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{allEmpty}})
	assert.True(t, pass, "both tags are empty")
	assert.Equal(t, []string{}, diff)

	oneFilled := makeManifest(`
data:
  config.json: |
    {"servers":[{"tag":null},{"tag":"value2"}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{oneFilled}})
	assert.False(t, pass, "one tag is null, so the assertion fails")
}
