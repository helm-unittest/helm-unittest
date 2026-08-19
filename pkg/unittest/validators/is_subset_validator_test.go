package validators_test

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestIsSubset = `
a:
  b:
    c: hello world
    d: foo bar
    x: baz
`

func TestIsSubsetValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar", "x": "baz"}}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "hello bar", "c": "hello world"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	log.SetLevel(log.DebugLevel)

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: bar bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestIsSubset)
	extraDoc := `
a:
  b:
    c: hello world
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: manifests,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	d: foo bar",
		"Actual:",
		"	c: hello world",
	}, diff)
}

func TestIsSubsetValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest1 := makeManifest(docToTestIsSubset)
	manifests := []common.K8sManifest{manifest1, manifest1}

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: manifests,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected NOT to contain:",
		"	d: foo bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorWhenNotAnObject(t *testing.T) {
	manifestDocNotObject := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotObject)

	validator := IsSubsetValidator{Path: "a.b.c", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Error:",
		"	expect 'a.b.c' to be an object, got:",
		"	hello world",
	}, diff)
}

func TestIsSubsetValidatorWhenNotAnObjectFailFast(t *testing.T) {
	manifestDocNotObject := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotObject)

	validator := IsSubsetValidator{Path: "a.b.c", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Error:",
		"	expect 'a.b.c' to be an object, got:",
		"	hello world",
	}, diff)
}

func TestIsSubsetValidatorWhenInvalidPath(t *testing.T) {
	manifest := makeManifest("a: error")

	validator := IsSubsetValidator{Path: "a[b]", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestIsSubsetValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest("a: error")

	validator := IsSubsetValidator{Path: "a[5]", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a[5]",
	}, diff)
}

func TestIsSubsetValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest("a: error")

	validator := IsSubsetValidator{Path: "a[5]", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorWhenUnknownPathFailFast(t *testing.T) {
	manifest := makeManifest("a: error")

	validator := IsSubsetValidator{Path: "a[5]", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a[5]",
	}, diff)
}

func TestIsSubsetValidatorWhenInvalidPathFailFast(t *testing.T) {
	manifest := makeManifest("a: error")

	validator := IsSubsetValidator{Path: "a[b]", Content: common.K8sManifest{"d": "foo bar"}}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestIsSubsetValidatorWhenFailFast(t *testing.T) {
	manifest := makeManifest(docToTestIsSubset)

	log.SetLevel(log.DebugLevel)

	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	e: bar bar",
		"Actual:",
		"	c: hello world",
		"	d: foo bar",
		"	x: baz",
	}, diff)
}

func TestIsSubsetValidatorWhenNoManifestFail(t *testing.T) {
	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\ta.b",
		"Expected to contain:",
		"\te: bar bar",
		"Actual:",
		"\tno manifest found",
	}, diff)
}

func TestIsSubsetValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := IsSubsetValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar bar"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorParsedJSONTopLevelKey(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"debug":false,"server":{"port":8080,"host":"0.0.0.0"}}
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      map[string]any{"debug": false},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorParsedYAMLViaInnerPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    server:
      port: 8080
      host: 0.0.0.0
    debug: false
`)

	validator := IsSubsetValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "server"},
		Content:      map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// The parsed JSON number must compare equal to the test file's int, which is
// what ParseOptions' number normalization guarantees.
func TestIsSubsetValidatorParsedJSONNumberMatchesIntExpectation(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server"},
		Content:      map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestIsSubsetValidatorParsedNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"debug":false}
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      map[string]any{"debug": true},
	}

	pass, _ := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
}

// isSubset requires an object. A parsed top-level array must report the
// existing "to be an object" error rather than silently passing.
func TestIsSubsetValidatorParsedArrayIsNotAnObject(t *testing.T) {
	manifest := makeManifest(`
data:
  list.json: |
    [{"a":1}]
`)

	validator := IsSubsetValidator{
		Path:         `data["list.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      map[string]any{"a": 1},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), "to be an object")
}

func TestIsSubsetValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"broken": }
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      map[string]any{"broken": 1},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), `unable to parse path 'data["config.json"]' as json`)
}

func TestIsSubsetValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.missing"},
		Content:      map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, `data["config.json"]`)
	assert.Contains(t, joined, "server.missing")
}

// innerPath can match several nodes; every one must satisfy the subset check.
func TestIsSubsetValidatorParsedFanOut(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080,"tls":true},{"port":8080,"tls":false}]}
`)

	validator := IsSubsetValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers[*]"},
		Content:      map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass, "both nodes contain port 8080")
	assert.Equal(t, []string{}, diff)

	mixed := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":8080},{"port":9090}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{mixed}})
	assert.False(t, pass, "one node has a different port, so the subset check fails")
}
