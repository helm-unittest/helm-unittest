package validators_test

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestExistsValidatorWhenOk(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := ExistsValidator{Path: "a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenArrayOk(t *testing.T) {
	doc := `
a:
  - b
`
	manifest := makeManifest(doc)

	v := ExistsValidator{Path: "a[0]"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenNegativeAndOk(t *testing.T) {
	doc := "a: 0"
	manifest := makeManifest(doc)

	v := ExistsValidator{Path: "b"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorWhenFail(t *testing.T) {
	doc := "a: A"
	manifest := makeManifest(doc)

	log.SetLevel(log.DebugLevel)

	v := ExistsValidator{Path: "b"}
	pass, diff := v.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	b expected to exists",
	}, diff)
}

func TestExistsValidatorWhenNegativeAndFail(t *testing.T) {
	doc := "a:"
	manifest := makeManifest(doc)

	v := ExistsValidator{Path: "a"}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a expected to NOT exists",
	}, diff)
}

func TestExistsValidatorWhenInvalidPath(t *testing.T) {
	doc := "x:"
	manifest := makeManifest(doc)

	validator := ExistsValidator{Path: "x[b]"}
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

func TestExistsValidatorWhenInvalidPathFailFast(t *testing.T) {
	doc := "x:"
	manifest := makeManifest(doc)
	secondManifest := makeManifest(doc)

	validator := ExistsValidator{Path: "x[b]"}
	pass, diff := validator.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, secondManifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [b] before position 4: non-integer array index",
	}, diff)
}

func TestExistsValidatorWhenNoManifestFail(t *testing.T) {
	validator := ExistsValidator{Path: "x"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\tx expected to exists",
	}, diff)
}

func TestExistsValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := ExistsValidator{Path: "x"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestExistsValidatorParsedInnerPathExists(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"tls":{"enabled":true}}}
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.tls.enabled"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	// Without parsing actually running, the outer path's raw string content is
	// always non-empty, so any innerPath would trivially "exist". Checking a
	// sibling key that is genuinely absent inside the same embedded JSON proves
	// the innerPath lookup is really operating on the parsed structure.
	absentSibling := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.tls.missingKey"},
	}
	pass, _ = absentSibling.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.False(t, pass, "server.tls.missingKey is absent from the embedded JSON")
}

// The key point of this validator plus parse: a key MISSING inside the embedded
// JSON must fail, even though the outer path resolves fine.
func TestExistsValidatorParsedInnerPathMissingFails(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.tls.enabled"},
	}

	pass, _ := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass,
		"the outer path resolves, but the innerPath does not, so it must not exist")
}

// Covers the notExists / isNull antonym path.
func TestExistsValidatorParsedInnerPathMissingPassesNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server.tls.enabled"},
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// A present-but-false value still EXISTS. This distinguishes existence from
// truthiness, and only holds if parsing actually ran.
func TestExistsValidatorParsedFalseValueStillExists(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    server:
      tls:
        enabled: false
`)

	validator := ExistsValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "server.tls.enabled"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	pass, _ = validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.False(t, pass, "the key exists, so notExists must fail")

	// Without parsing actually running, the outer path's raw string is always
	// non-empty, so any innerPath would trivially "exist" regardless of whether
	// the key is really there. A sibling key absent from the embedded YAML
	// proves the lookup operates on the parsed structure, not the raw string.
	absentSibling := ExistsValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "server.tls.certPath"},
	}
	pass, _ = absentSibling.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.False(t, pass, "server.tls.certPath is absent from the embedded YAML")
}

func TestExistsValidatorParsedArrayIndex(t *testing.T) {
	manifest := makeManifest(`
data:
  dashboard.json: |
    [{"title":"Latency"}]
`)

	present := ExistsValidator{
		Path:         `data["dashboard.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "[0].title"},
	}
	pass, diff := present.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)

	absent := ExistsValidator{
		Path:         `data["dashboard.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "[5].title"},
	}
	pass, _ = absent.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.False(t, pass, "index 5 is out of range, so it must not exist")
}

func TestExistsValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    not json at all {
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "anything"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}

// Malformed content must fail even under a negative assertion: a parse error is
// a real error, not evidence of absence.
func TestExistsValidatorParseFailureFailsEvenWhenNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    not json at all {
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "anything"},
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), "unable to parse path")
}

func TestExistsValidatorParsedWithoutInnerPathChecksParsedValue(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"a":1}
`)

	validator := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass, "the parsed document itself exists")
	assert.Equal(t, []string{}, diff)

	// The outer path resolves to a non-empty string either way, so the
	// positive case above would pass even without parsing actually running.
	// Malformed content, with no innerPath, only fails if resolveActuals is
	// really parsing the value: proof that parsing runs even when innerPath
	// is omitted.
	malformed := makeManifest(`
data:
  config.json: |
    not json at all {
`)
	invalid := ExistsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
	}
	pass, diff = invalid.Validate(&ValidateContext{Docs: []common.K8sManifest{malformed}})
	assert.False(t, pass, "malformed content must fail to parse even without innerPath")
	assert.Contains(t, strings.Join(diff, "\n"), "unable to parse path")
}
