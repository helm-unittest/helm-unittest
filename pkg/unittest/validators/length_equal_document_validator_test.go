package validators_test

import (
	"strings"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	"github.com/stretchr/testify/assert"
)

var (
	testDocLengthEqual1 = `
spec:
  tls:
   - hosts:
      - a.example.com
      - b.example.com
     secretName: example.com
`
	testDocLengthEqual2 = `
spec:
  tls:
   - hosts:
      - a.example.com
     secretName: a.example.com
   - hosts:
      - b.example.com
     secretName: b.example.com
`
	testDocLengthEqual3_Success = `
spec:
  tls:
   - hosts:
      - a.example.com
     secretName: a.example.com
   - hosts:
      - b.example.com
     secretName: b.example.com
  rules:
   - host: a.example.com
   - host: b.example.com
`
	testDocLengthEqual3_Fail = `
spec:
  tls:
   - hosts:
      - a.example.com
     secretName: a.example.com
   - hosts:
      - b.example.com
     secretName: b.example.com
  rules:
   - host: a.example.com
`
	testDocLengthEqual0_Success = `
spec:
  volumes:
`
)

func TestLengthEqualDocumentsValidatorOk_Single(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual1)
	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorOk_Single2(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual2)
	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorNegativeOk_Single(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual1)
	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorNegativeOk_SingleNoPath(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual1)
	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.ssl",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorOk_Multi(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Success)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls", "spec.rules"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorNegative_MultiNoPath(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Success)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.ssl", "spec.rules"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorNegativeFail_Multi(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Success)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls", "spec.rules"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0",
		"Path:\tspec.tls", "Expected NOT to match count:", "\t-1", "Actual:", "\t2"}, diff)
}

func TestLengthEqualDocumentsValidatorFail_Single(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual2)
	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0", "Path:\tspec.tls",
		"Expected to match count:", "\t1", "Actual:", "\t2"}, diff)
}

func TestLengthEqualDocumentsValidatorNegativeFail_Single(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual2)
	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0", "Path:\tspec.tls",
		"Expected NOT to match count:", "\t2", "Actual:", "\t2"}, diff)
}

func TestLengthEqualDocumentsValidatorFail_Multi(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls", "spec.rules"},
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:\t0", "Path:\tspec.tls",
		"Expected to match count:", "\t-1", "Actual:", "\t2"}, diff)
}

func TestLengthEqualDocumentsValidatorWhenPathAndNoCount(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	validator := LengthEqualDocumentsValidator{
		Path: "spec.tls",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"Error:", "\t'count' field must be set if 'path' is used"}, diff)
}

func TestLengthEqualDocumentsValidatorWhenPathAndNegativeCount(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	count := -24
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"Error:", "\t'count' field must be set if 'path' is used"}, diff)
}

func TestLengthEqualDocumentsValidatorWhenBadConfig(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	count := 2
	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls"},
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"Error:", "\t'paths' couldn't be used with 'path'"}, diff)
}

func TestLengthEqualDocumentsValidatorOk_Empty(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual0_Success)
	count := 0
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.volumes",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorOk_WhenNegative(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual0_Success)
	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.volumes",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorNoManifestFail(t *testing.T) {
	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Path:\t",
		"Expected to match count:",
		"\t",
		"Actual:",
		"\tno manifest found"}, diff)
}

func TestLengthEqualDocumentsValidatorNoManifestNegativeOk(t *testing.T) {
	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: &count,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorParsedSinglePath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":2},{"port":3}]}
`)

	count := 3
	validator := LengthEqualDocumentsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorParsedTopLevelArray(t *testing.T) {
	manifest := makeManifest(`
data:
  list.json: |
    ["a","b"]
`)

	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:         `data["list.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON},
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorParsedWrongCountFails(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1}]}
`)

	count := 3
	validator := LengthEqualDocumentsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"), "1",
		"the failure must report the decoded array length, proving parsing ran")
}

// parse applies to every path in Paths; the same innerPath is resolved within
// each, and the resulting counts are compared across them as before.
func TestLengthEqualDocumentsValidatorParsedMultiplePaths(t *testing.T) {
	manifest := makeManifest(`
data:
  a.json: |
    {"servers":[{"port":1},{"port":2}]}
  b.json: |
    {"servers":[{"port":3},{"port":4}]}
`)

	validator := LengthEqualDocumentsValidator{
		Paths:        []string{`data["a.json"]`, `data["b.json"]`},
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass, "both parsed documents have equal-length servers arrays")
	assert.Equal(t, []string{}, diff)
}

func TestLengthEqualDocumentsValidatorParsedMultiplePathsMismatchFails(t *testing.T) {
	manifest := makeManifest(`
data:
  a.json: |
    {"servers":[{"port":1},{"port":2}]}
  b.json: |
    {"servers":[{"port":3}]}
`)

	validator := LengthEqualDocumentsValidator{
		Paths:        []string{`data["a.json"]`, `data["b.json"]`},
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass, "the two parsed arrays differ in length")
	assert.Contains(t, strings.Join(diff, "\n"), "1",
		"the failure must report the decoded array length (1), proving parsing ran")
}

func TestLengthEqualDocumentsValidatorParsedYAML(t *testing.T) {
	manifest := makeManifest(`
data:
  config.yaml: |
    tags:
      - a
      - b
`)

	count := 2
	validator := LengthEqualDocumentsValidator{
		Path:         `data["config.yaml"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatYAML, InnerPath: "tags"},
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

// Covers the notLengthEqual antonym path.
func TestLengthEqualDocumentsValidatorParsedNegative(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {"servers":[{"port":1},{"port":2}]}
`)

	wrongCount := 5
	mismatch := LengthEqualDocumentsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Count:        &wrongCount,
	}
	pass, _ := mismatch.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass, "the length is not 5, so notLengthEqual holds")

	rightCount := 2
	matching := LengthEqualDocumentsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Count:        &rightCount,
	}
	pass, _ = matching.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.False(t, pass, "the length IS 2, so notLengthEqual must fail")
}

// lengthEqual needs an array. A parsed object must report the existing
// "is not array" error rather than being silently accepted.
//
// The "is not array" message only ever names the path, never the value, so a
// single path resolving to a parsed object is indistinguishable in the
// error text from the same path's raw (unparsed) string content - both are
// non-arrays and produce byte-identical output. To make this test actually
// depend on parsing having run, it pairs two paths sharing one innerPath:
// one whose parsed innerPath value is an array (must succeed, proving
// decode occurred) and one whose parsed innerPath value is an object (must
// fail "is not array"). Without parsing, both paths' raw string content is
// non-array, so the array-path would ALSO fail "is not array" - which this
// test asserts does not happen.
func TestLengthEqualDocumentsValidatorParsedNonArrayReportsError(t *testing.T) {
	manifest := makeManifest(`
data:
  arr.json: |
    {"x":[1,2,3]}
  obj.json: |
    {"x":{"y":1}}
`)

	validator := LengthEqualDocumentsValidator{
		Paths:        []string{`data["arr.json"]`, `data["obj.json"]`},
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "x"},
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	joined := strings.Join(diff, "\n")
	assert.Contains(t, joined, `data["obj.json"] is not array`,
		"the object-shaped innerPath value must fail the array check")
	assert.NotContains(t, joined, `data["arr.json"] is not array`,
		"the array-shaped innerPath value must have been decoded successfully, proving parsing ran")
}

func TestLengthEqualDocumentsValidatorParseFailureReportsPath(t *testing.T) {
	manifest := makeManifest(`
data:
  config.json: |
    {broken
`)

	count := 1
	validator := LengthEqualDocumentsValidator{
		Path:         `data["config.json"]`,
		ParseOptions: ParseOptions{Parse: ParseFormatJSON, InnerPath: "servers"},
		Count:        &count,
	}

	pass, diff := validator.Validate(&ValidateContext{Docs: []common.K8sManifest{manifest}})

	assert.False(t, pass)
	assert.Contains(t, strings.Join(diff, "\n"),
		`unable to parse path 'data["config.json"]' as json`)
}
