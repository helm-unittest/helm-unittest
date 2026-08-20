package validators_test

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestType = `
a:
  b:
    - c: 123
  e: |
    Line1
    Line2
`

func TestTypeValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestType)
	validator := IsTypeValidator{Path: "a.b[0].c", Type: "int"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorMultiLineWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestType)
	validator := IsTypeValidator{Path: "a.e", Type: "string"}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestType)

	validator := IsTypeValidator{Path: "a.b[0].c", Type: "string"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	log.SetLevel(log.DebugLevel)

	validator := IsTypeValidator{Path: "a.b[0]", Type: "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	map[string]interface {}",
	}, diff)
}

func TestTypeValidatorWhenFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	log.SetLevel(log.DebugLevel)

	validator := IsTypeValidator{Path: "a.b[0]", Type: "int"}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	map[string]interface {}",
	}, diff)
}

func TestTypeValidatorMultiManifestWhenFail(t *testing.T) {
	correctDoc := `
a:
  b:
    - c: "123"
`
	manifest1 := makeManifest(correctDoc)
	manifest2 := makeManifest(docToTestType)

	validator := IsTypeValidator{Path: "a.b[0].c", Type: "string"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b[0].c",
		"Expected to be of type:",
		"	string",
		"Actual:",
		"	int",
	}, diff)
}

func TestTypeValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	validator := IsTypeValidator{Path: "a.e", Type: "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.e",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	string",
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.e",
		"Expected to be of type:",
		"	int",
		"Actual:",
		"	string",
	}, diff)
}

func TestTypeValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.e", Type: "string"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.e",
		"Expected NOT to be of type:",
		"	string",
		"Actual:",
		"	string",
	}, diff)
}

func TestTypeValidatorWhenWrongPath(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.b[e]", Type: "int"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestTypeValidatorWhenWrongPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.b[e]", Type: "int"}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestTypeValidatorWhenUnkownPath(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.b[5]", Type: "string"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}

func TestTypeValidatorWhenUnkownPathNegative(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.b[5]", Type: "string"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestTypeValidatorWhenUnkownPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestType)

	v := IsTypeValidator{Path: "a.b[5]", Type: "string"}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}

func TestTypeValidatorWhenNoManifestFail(t *testing.T) {
	validator := IsTypeValidator{Path: "a.b[0]", Type: "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\ta.b[0]",
		"Expected to be of type:",
		"\tint",
		"Actual:",
		"\tno manifest found",
	}, diff)
}

func TestTypeValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := IsTypeValidator{Path: "a.b[0]", Type: "int"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// #794: reflect.TypeOf(nil) returns nil, so calling .String() on an
// unset YAML key used to crash the plugin. The validator must fail the
// assertion cleanly instead of panicking.
var docToTestTypeNil = `
a:
  nil_key:
  b:
    - c: 123
`

func TestTypeValidatorOnNilValueDoesNotPanic(t *testing.T) {
	manifest := makeManifest(docToTestTypeNil)

	validator := IsTypeValidator{Path: "a.nil_key", Type: "bool"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.NotEmpty(t, diff)
}

func TestTypeValidatorOnNilValueNegativeStillDoesNotPanic(t *testing.T) {
	manifest := makeManifest(docToTestTypeNil)

	validator := IsTypeValidator{Path: "a.nil_key", Type: "bool"}
	pass, _ := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
}

// The bug class this feature exists to catch: a hand-written template dropped
// the quotes, so a value that should be a JSON string is a JSON number.
// Unparsed, the actual is one opaque string and this would wrongly pass.
func TestIsTypeValidatorParsedDetectsUnquotedNumber(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"fromStr":8080}
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "fromStr"},
		Type:         "string",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass, "an unquoted JSON number must not satisfy type string")
	assert.Contains(t, strings.Join(diff, "\n"), "int",
		"the reported actual type must be the decoded int, proving parsing ran")
}

func TestIsTypeValidatorParsedQuotedNumberIsString(t *testing.T) {
	// fromNum is a genuine numeric sibling. Unparsed, the actual for either
	// innerPath is the same opaque raw-string blob matched by Path, so
	// asserting Type "string" on fromStr alone would pass whether or not
	// parsing is wired in - it is not a useful regression check by itself.
	// The fromNum/int companion assertion below is what actually depends on
	// parsing: only a real decode turns that field into an int, so it fails
	// without the wiring and passes with it.
	manifest := makeManifest(`
data:
  config.json: |
    {"fromStr":"8080","fromNum":8080}
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "fromStr"},
		Type:         "string",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	asInt := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "fromNum"},
		Type:         "int",
	}
	pass, diff = asInt.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass, "the numeric sibling must decode to int, proving parsing ran")
	assert.Equal(t, []string{}, diff)
}

// Normalization means an integer JSON literal reports int, not float64. The
// float64 assertion below only means something paired with the preceding int
// assertion on the same manifest: on its own it would also "pass" against the
// raw unparsed JSON text (which is a string, so also not float64), so it must
// stay coupled to the positive check rather than become a standalone case.
func TestIsTypeValidatorParsedIntegerReportsInt(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"port":8080}
`)

	asInt := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Type:         "int",
	}
	pass, diff := asInt.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	asFloat := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Type:         "float64",
	}
	pass, _ = asFloat.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.False(t, pass, "integer syntax must not report float64")
}

// Both cases parse a single scalar to its expected native type: a decimal
// JSON literal must report float64 (not int), and a YAML bool literal must
// report bool.
func TestIsTypeValidatorParsedScalarTypes(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		content      string
		parseFormat  string
		innerPath    string
		expectedType string
	}{
		{
			// 1.5 (rather than an integral decimal like 1.0) avoids a known
			// unrelated quirk of innerPath resolution: it re-encodes the parsed
			// value to YAML and back (via GetValueOfSetPath) to apply the JSONPath
			// lookup, and the YAML library renders an integral float without a
			// fractional part (e.g. "1"), which then decodes back as int. A
			// non-integral decimal round-trips through YAML without ambiguity and
			// still exercises "decimal-point syntax reports float64".
			name:         "decimal literal reports float64",
			fileName:     "config.json",
			content:      `{"ratio":1.5}`,
			parseFormat:  ParseFormatJSON,
			innerPath:    "ratio",
			expectedType: "float64",
		},
		{
			name:         "yaml bool",
			fileName:     "config.yaml",
			content:      "debug: true",
			parseFormat:  ParseFormatYAML,
			innerPath:    "debug",
			expectedType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := makeManifest("data:\n  " + tt.fileName + ": |\n    " + tt.content + "\n")

			validator := IsTypeValidator{
				Path:         `data["` + tt.fileName + `"]`,
				ParseOptions: ParseOptions{Parse: tt.parseFormat, InnerPath: tt.innerPath},
				Type:         tt.expectedType,
			}

			pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

			assert.True(t, pass)
			assert.Equal(t, []string{}, diff)
		})
	}
}

func TestIsTypeValidatorParsedObjectAndArray(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080},"tags":["a"]}
`)

	object := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server"},
		Type:         "map[string]interface {}",
	}
	pass, diff := object.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	array := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "tags"},
		Type:         "[]interface {}",
	}
	pass, diff = array.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// Covers the isNotType antonym path.
func TestIsTypeValidatorParsedNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"port":8080}
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Type:         "string",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass, "the parsed value is int, not string, so isNotType holds")
	assert.Equal(t, []string{}, diff)

	// And the matching type must FAIL the negative form, proving the type
	// comparison really ran against the parsed value.
	matching := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Type:         "int",
	}
	pass, _ = matching.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.False(t, pass)
}

func TestIsTypeValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {broken
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "port"},
		Type:         "int",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}

func TestIsTypeValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"a":1}
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "missing"},
		Type:         "int",
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

// innerPath may match several nodes; every one must have the expected type.
func TestIsTypeValidatorParsedFanOut(t *testing.T) {
	allInts := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":2}]}
`)

	validator := IsTypeValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Type:         "int",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{allInts}})
	assert.True(t, pass, "both ports are ints")
	assert.Equal(t, []string{}, diff)

	mixed := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":"2"}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{mixed}})
	assert.False(t, pass, "one port is a quoted string, so the assertion fails")
}
