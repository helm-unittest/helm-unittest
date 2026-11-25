package validators_test

import (
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"

	"github.com/stretchr/testify/assert"
)

var docToTestContains = `
a:
  b:
    - c: hello world
    - d: foo bar
    - e: bar
    - e: bar
`

var docToTestContains2 = `
a:
  b:
    - d: foo bar
`
var docToTestContains3 = `
a:
  b:
    - d:
`

func TestContainsValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenEmptyManifestFail(t *testing.T) {
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0",
		"Path:\ta.b",
		"Expected to contain:",
		"\t- d: foo bar",
		"Actual:", "\tno manifest found"}, diff)
}

func TestContainsValidatorWhenEmptyManifestNegativeOk(t *testing.T) {
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenOkWithMultiValues(t *testing.T) {

	var multiAssertToTestContains = `
a:
  - d: foo bar
  - d: foo bar
`

	manifest := makeManifest(multiAssertToTestContains)

	validator := ContainsValidator{
		Path:    "$.*",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMultiManifestContainsValidatorWhenOk(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifest2 := makeManifest(docToTestContains2)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest1, manifest2},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithValueOnlyWhenOk(t *testing.T) {
	docToTestContainsValueOnly := `
a:
  b:
    - VALUE1
    - VALUE2
`
	manifest := makeManifest(docToTestContainsValueOnly)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: "VALUE1",
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithValueOnlyAndAnyEnabledWhenOk(t *testing.T) {
	docToTestContainsValueOnly := `
a:
  b:
    - VALUE1
    - VALUE2
`
	manifest := makeManifest(docToTestContainsValueOnly)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: "VALUE1",
		Count:   nil,
		Any:     true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithAnyWhenOk(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"name": "VALUE1"},
		Count:   nil,
		Any:     true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWithAnyWhenNotFoundOk(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	// Enable debug logging
	log.SetLevel(log.DebugLevel)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"name": "VALUE3"},
		Count:   nil,
		Any:     true,
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
		"	- name: VALUE3",
		"Actual:",
		"	- name: VALUE1",
		"	  value: bla",
		"	- name: VALUE2",
		"	  value: bla2",
	}, diff)
}

func TestContainsValidatorWithAnyWhenNotFoundAndMultiManifest(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	// Enable debug logging
	log.SetLevel(log.DebugLevel)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"name": "VALUE3"},
		Count:   nil,
		Any:     true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest, manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- name: VALUE3",
		"Actual:",
		"	- name: VALUE1",
		"	  value: bla",
		"	- name: VALUE2",
		"	  value: bla2",
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- name: VALUE3",
		"Actual:",
		"	- name: VALUE1",
		"	  value: bla",
		"	- name: VALUE2",
		"	  value: bla2",
	}, diff)
}

func TestContainsValidatorWithMultiManifestAndFailfast(t *testing.T) {
	docToTestContainsAny := `
a:
  b:
    - name: VALUE1
      value: bla
    - name: VALUE2
      value: bla2
`
	manifest := makeManifest(docToTestContainsAny)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"name": "VALUE3"},
		Count:   nil,
		Any:     true,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest, manifest},
		FailFast: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- name: VALUE3",
		"Actual:",
		"	- name: VALUE1",
		"	  value: bla",
		"	- name: VALUE2",
		"	  value: bla2",
	}, diff)
}

func TestContainsValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "hello bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar bar"},
		Count:   nil,
		Any:     false,
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
		"	- e: bar bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	extraDoc := `
a:
  b:
    - c: hello world
`
	manifest2 := makeManifest(extraDoc)
	manifests := []common.K8sManifest{manifest1, manifest2}

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
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
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
	}, diff)
}

func TestContainsValidatorMultiManifestWhenBothFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifests := []common.K8sManifest{manifest1, manifest1}

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "foo bar"},
		Count:   nil,
		Any:     false,
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
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
		"DocumentIndex:	1",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected to contain:",
		"	- e: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
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
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorMultiDocsWhenNegativeAndFail(t *testing.T) {
	manifest1 := makeManifest(docToTestContains)
	manifest2 := makeManifest(docToTestContains3)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest1, manifest2},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Path:	a.b",
		"Expected NOT to contain:",
		"	- d: foo bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestMatchContainsValidatorWhenNotAnArray(t *testing.T) {
	manifestDocNotArray := `
a:
  b:
    c: hello world
    d: foo bar
`
	manifest := makeManifest(manifestDocNotArray)

	validator := ContainsValidator{
		Path:    "a.b",
		Content: common.K8sManifest{"d": "foo bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Error:",
		"	expect 'a.b' to be an array, got:",
		"	c: hello world",
		"	d: foo bar",
	}, diff)
}

func TestContainsValidatorWhenInvalidParameter(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b[e]",
		Content: common.K8sManifest{"e": "bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestContainsValidatorWhenInvalidParameterFailfast(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b[e]",
		Content: common.K8sManifest{"e": "bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest, manifest},
		FailFast: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	invalid array index [e] before position 6: non-integer array index",
	}, diff)
}

func TestContainsValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b[5]",
		Content: common.K8sManifest{"e": "bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}

func TestContainsValidatorWhenUnknownPathFailfast(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b[5]",
		Content: common.K8sManifest{"e": "bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest, manifest},
		FailFast: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a.b[5]",
	}, diff)
}

func TestContainsValidatorWhenUnknownPathNegative(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	validator := ContainsValidator{
		Path:    "a.b[5]",
		Content: common.K8sManifest{"e": "bar"},
		Count:   nil,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := 2
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar"},
		Count:   &counter,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorInverseWhenNotMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := 1
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar"},
		Count:   &counter,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorWhenNotMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := 1
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"e": "bar"},
		Count:   &counter,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"ValuesIndex:	0",
		"Error:",
		"	expect count 1 in 'a.b' to be in array, got 2:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorWhenNotFoundMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := 1
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"f": "bar"},
		Count:   &counter,
		Any:     false,
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
		"	- f: bar",
		"Actual:",
		"	- c: hello world",
		"	- d: foo bar",
		"	- e: bar",
		"	- e: bar",
	}, diff)
}

func TestContainsValidatorInverseWhenNotFoundMultipleTimesInArray(t *testing.T) {
	manifest := makeManifest(docToTestContains)

	counter := 1
	validator := ContainsValidator{
		Path:    "a.b",
		Content: map[string]any{"f": "bar"},
		Count:   &counter,
		Any:     false,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorParsedJSONArrayViaInnerPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"host":"a","port":1},{"host":"b","port":2}]}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Content:      map[string]any{"host": "a", "port": 1},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorParsedTopLevelArray(t *testing.T) {
	manifest := makeManifest(`
data:
  list.json: |
    ["alpha","beta"]
`)

	validator := ContainsValidator{
		Path:         `data["list.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      "beta",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// Numbers in the parsed content must compare equal to the test file's ints,
// which is what ParseOptions' number normalization guarantees.
func TestContainsValidatorParsedNumbersMatchIntExpectations(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"ports":[8080,9090]}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "ports"},
		Content:      8080,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorParsedWithCount(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"tags":["a","a","b"]}
`)

	count := 2
	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "tags"},
		Content:      "a",
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorParsedWithAnySubset(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"host":"a","port":1,"extra":true}]}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Content:      map[string]any{"host": "a"},
		Any:          true,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsValidatorParsedYAMLNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    tags:
      - a
      - b
`)

	validator := ContainsValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "tags"},
		Content:      "zzz",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// contains requires an array. A parsed object must report the existing
// "to be an array" error, showing the DECODED object as the actual, which is
// what distinguishes this from the unparsed raw-string case.
func TestContainsValidatorParsedNonArrayReportsError(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"server":{"port":8080}}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "server"},
		Content:      map[string]any{"port": 8080},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, "to be an array")
	assert.Contains(t, joined, "port: 8080",
		"the actual must be the decoded object, proving parsing ran")
}

func TestContainsValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"broken": }
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Content:      "x",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), `unable to parse path 'data["config.json"]' as json`)
}

func TestContainsValidatorParsedUnmatchedInnerPathNamesBothPaths(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"host":"a"}]}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "missing"},
		Content:      map[string]any{"host": "a"},
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

// innerPath may match several arrays; every one must contain the content.
func TestContainsValidatorParsedFanOut(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"groups":[{"tags":["x","y"]},{"tags":["x","z"]}]}
`)

	validator := ContainsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "groups[*].tags"},
		Content:      "x",
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})
	assert.True(t, pass, "both tag arrays contain x")
	assert.Equal(t, []string{}, diff)

	mixed := makeManifest(`
data:
  config.json: |
    {"groups":[{"tags":["x"]},{"tags":["q"]}]}
`)

	pass, _ = validator.Validate(&ValidateContext{Docs: []common.K8sManifest{mixed}})
	assert.False(t, pass, "one array lacks x, so the assertion fails")
}
