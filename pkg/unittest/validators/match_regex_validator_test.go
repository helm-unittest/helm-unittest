package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestMatchRegex = `
a:
  b:
    - c: hello world
e: |
  aaa
  bbb
`

var docToTestMatchRegexWithBase64 = `
a: aGVsbG8gd29ybGQ=
b: YWFhCmJiYgo=
`

func TestMatchRegexValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^hello", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenMultiLineOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"e", "bbb", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWithBase64WhenNOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^hello", true}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"DocumentIndex:	0", "Error:", "	unable to decode base64 expected content hello world"}, diff)
}

func TestMatchRegexValidatorWithBase64WhenOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexWithBase64)

	validator := MatchRegexValidator{"a", "^hello", true}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenMultiLineWithBase64Ok(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexWithBase64)

	validator := MatchRegexValidator{"b", "bbb", true}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^foo", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenRegexCompileFail(t *testing.T) {
	manifest := common.K8sManifest{"a": "A"}

	validator := MatchRegexValidator{"a", "+", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	error parsing regexp: missing argument to repetition operator: `+`",
	}, diff)
}

func TestMatchRegexValidatorWhenNotString(t *testing.T) {
	manifest := common.K8sManifest{"a": 123.456}

	validator := MatchRegexValidator{"a", "^foo", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	expect 'a' to be a string, got:",
		"	123.456",
	}, diff)
}

func TestMatchRegexValidatorWhenMatchFail(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	log.SetLevel(log.DebugLevel)

	validator := MatchRegexValidator{"a.b[0].c", "^foo", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0].c",
		"Expected to match:",
		"	^foo",
		"Actual:",
		"	hello world",
	}, diff)
}

func TestMatchRegexValidatorWhenNegativeAndMatchFail(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^hello", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Path:	a.b[0].c",
		"Expected NOT to match:",
		"	^hello",
		"Actual:",
		"	hello world",
	}, diff)
}

func TestMatchRegexValidatorWhenNoPattern(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	expected field 'pattern' to be filled",
	}, diff)
}

func TestMatchRegexValidatorWhenErrorGetValueOfSetPath(t *testing.T) {
	manifest := makeManifest("a.b.d::error")

	validator := MatchRegexValidator{"a.[b]", "^hello", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	child name missing at position 2, following \"a.\"",
	}, diff)
}

func TestMatchRegexValidatorWhenUnknownPath(t *testing.T) {
	manifest := makeManifest("a.b.d::error")

	validator := MatchRegexValidator{"a[2]", "^hello", false}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	unknown path a[2]",
	}, diff)
}
