package validators_test

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestEqual = `
a:
  b:
    - c: 123
  e: |
    Line1 
    Line2
`

var docToTestEqualWithBase64 = `
a: MTIz
b: TGluZTEgCkxpbmUyCg==
`

var docToTestEqualMultiplePaths = `
a:
  b: 1
  c: 1
`

func TestEqualValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{Path: "a.b[0].c", Value: 123, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorMultiLineWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{Path: "a.e", Value: "Line1\nLine2\n", DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWithBase64WhenNOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)
	validator := EqualValidator{Path: "a.e", Value: "Line1\nLine2\n", DecodeBase64: true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:	0", "ValuesIndex:	0", "Error:", "	unable to decode base64 expected content Line1 ", "	Line2"}, diff)
}

func TestEqualValidatorWithBase64WhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqualWithBase64)
	validator := EqualValidator{Path: "a", Value: "123", DecodeBase64: true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorMultiLineWithBase64WhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqualWithBase64)
	validator := EqualValidator{Path: "b", Value: "Line1\nLine2\n", DecodeBase64: true}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	validator := EqualValidator{Path: "a.b[0].c", Value: 321, DecodeBase64: false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	log.SetLevel(log.DebugLevel)

	validator := EqualValidator{
		Path:         "a.b[0]",
		Value:        map[any]any{"d": 321},
		DecodeBase64: false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	d: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-d: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorMultiManifestWhenFail(t *testing.T) {
	correctDoc := `
a:
  b:
    - c: 321
`
	manifest1 := makeManifest(correctDoc)
	manifest2 := makeManifest(docToTestEqual)

	validator := EqualValidator{
		Path:         "a.b[0]",
		Value:        map[string]any{"c": 321},
		DecodeBase64: false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	validator := EqualValidator{
		Path:         "a.b[0]",
		Value:        map[string]any{"c": 321},
		DecodeBase64: false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected to equal:",
		"	c: 321",
		"Actual:",
		"	c: 123",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-c: 321",
		"	+c: 123",
	}, diff)
}

func TestEqualValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{Path: "a.b[0]", Value: map[string]any{"c": 123}, DecodeBase64: false}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b[0]",
		"Expected NOT to equal:",
		"	c: 123",
	}, diff)
}

func TestEqualValidatorWhenWrongPath(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{Path: "a.b[e]", Value: map[string]int{"d": 321}, DecodeBase64: false}
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

func TestEqualValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{Path: "a.b[5]", Value: map[string]int{"d": 321}, DecodeBase64: false}
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

func TestEqualValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{Path: "a.b[5]", Value: map[string]int{"d": 321}, DecodeBase64: false}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWhenUnknownPathFailFast(t *testing.T) {
	manifest := makeManifest(docToTestEqual)

	v := EqualValidator{Path: "a.b[5]", Value: map[string]int{"d": 321}, DecodeBase64: false}
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

func TestEqualValidatorWhenOkWithMultiplePaths(t *testing.T) {
	manifest := makeManifest(docToTestEqualMultiplePaths)
	validator := EqualValidator{Path: "a.*", Value: 1, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorWithMultiplePathsFailFast(t *testing.T) {
	manifest := makeManifest(docToTestEqualMultiplePaths)
	validator := EqualValidator{Path: "a.*", Value: 2, DecodeBase64: true}

	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"ValuesIndex:\t0",
		"Path:\ta.*",
		"Expected to equal:",
		"\t2",
		"Actual:",
		"\t1",
		"Diff:",
		"\t--- Expected",
		"\t+++ Actual",
		"\t@@ -1,2 +1,2 @@",
		"\t-2",
		"\t+1"}, diff)
}

func TestEqualValidatorWhenNoManifestFail(t *testing.T) {
	validator := EqualValidator{Path: "a.b[0].c", Value: 123, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\ta.b[0].c",
		"Expected to equal:",
		"\t123",
		"Actual:",
		"\tno manifest found",
		"Diff:",
		"\t--- Expected",
		"\t+++ Actual",
		"\t@@ -1,2 +1,2 @@",
		"\t-123",
		"\t+no manifest found"}, diff)
}

func TestEqualValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := EqualValidator{Path: "a.b[0].c", Value: 123, DecodeBase64: false}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// This is the canary for the whole design: without number normalization the
// test-file int 8080 would not equal the parsed JSON value, because
// encoding/json decodes numbers as float64.
func TestEqualValidatorParsedJSONNumberMatchesIntExpectation(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"},
		Value:        8080,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorParsedJSONIgnoresKeyOrdering(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"zebra":1,"alpha":2,"middle":3}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Value: map[string]any{
			"alpha":  2,
			"middle": 3,
			"zebra":  1,
		},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorParsedYAMLDeepLeaf(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    server:
      tls:
        enabled: true
`)

	validator := EqualValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "server.tls.enabled"},
		Value:        true,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorParsedArrayElement(t *testing.T) {
	manifest := makeManifest(`
data:
  dashboard.json: |
    [{"title":"Latency"},{"title":"QPS"}]
`)

	validator := EqualValidator{
		Path:         `data["dashboard.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "[1].title"},
		Value:        "QPS",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// innerPath, like path, may match several nodes; all of them must satisfy the
// comparison.
func TestEqualValidatorParsedFanOutRequiresAllToMatch(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080},{"port":8080}]}
`)

	allMatch := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        8080,
	}

	pass, diff := allMatch.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	mixed := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080},{"port":9090}]}
`)

	pass, _ = allMatch.Validate(&ValidateContext{Docs: []common.K8sManifest{mixed}})
	assert.False(t, pass, "a single non-matching node must fail the assertion")
}

func TestEqualValidatorParseComposesWithDecodeBase64(t *testing.T) {
	// base64 of {"server":{"port":8080}}
	manifest := makeManifest(`
data:
  config: eyJzZXJ2ZXIiOnsicG9ydCI6ODA4MH19
`)

	validator := EqualValidator{
		Path:         "data.config",
		DecodeBase64: true,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"},
		Value:        8080,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualValidatorParsedNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.port"},
		Value:        9090,
	}

	pass, _ := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
}

func TestEqualValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"port": 8080,}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Value:        map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), `unable to parse path 'data["config.json"]' as json`)
}

func TestEqualValidatorParseOnNonStringReportsType(t *testing.T) {
	manifest := makeManifest(`
spec:
  replicas: 3
`)

	validator := EqualValidator{
		Path:         "spec.replicas",
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Value:        3,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		"expect 'spec.replicas' to be a string to parse as json, got int")
}

// An innerPath that matches nothing behaves like an unknown path: a failure
// normally, a pass under a negative assertion.
func TestEqualValidatorParsedUnmatchedInnerPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.missing"},
		Value:        1,
	}

	pass, _ := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.False(t, pass)

	pass, _ = validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass)
}

// Two distinct failing nodes in a parse fan-out must be distinguishable in the
// output, the same way a plain multi-match path distinguishes them.
func TestEqualValidatorParsedFanOutReportsDistinctIndices(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":9090},{"port":7070}]}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        8080,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, "ValuesIndex:\t0")
	assert.Contains(t, joined, "ValuesIndex:\t1")
}

func TestEqualValidatorParsedFanOutFailFastReportsOneFailure(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":2},{"port":3}]}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        9999,
	}

	_, failFastDiff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		FailFast: true,
	})
	_, fullDiff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.Less(t, len(failFastDiff), len(fullDiff),
		"FailFast must stop after the first failing parsed node")
}

// Negation composes with fan-out per node: every node must differ from Value.
func TestEqualValidatorParsedFanOutNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":2}]}
`)

	noneMatch := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        9999,
	}

	pass, _ := noneMatch.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass, "no node equals the value, so notEqual holds")

	oneMatches := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*].port"},
		Value:        1,
	}

	pass, _ = oneMatches.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.False(t, pass, "one node equals the value, so notEqual fails")
}

func TestEqualValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := EqualValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.missing"},
		Value:        1,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, `data["config.json"]`)
	assert.Contains(t, joined, "server.missing")
}
