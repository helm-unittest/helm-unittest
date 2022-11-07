package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
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

func TestMatchRegexValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^hello"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenMultiLineOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"e", "bbb"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^foo"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexValidatorWhenRegexCompileFail(t *testing.T) {
	manifest := common.K8sManifest{"a": "A"}

	validator := MatchRegexValidator{"a", "+"}
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

	validator := MatchRegexValidator{"a", "^foo"}
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

	validator := MatchRegexValidator{"a.b[0].c", "^foo"}
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

	validator := MatchRegexValidator{"a.b[0].c", "^hello"}
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

func TestMatchRegexValidatorWhenInvalidIndex(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", "^hello"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:  []common.K8sManifest{manifest},
		Index: 2,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	documentIndex 2 out of range",
	}, diff)
}

func TestMatchRegexValidatorWhenNoPattern(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexValidator{"a.b[0].c", ""}
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

	validator := MatchRegexValidator{"a.b[0].c", "^hello"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:	0",
		"Error:",
		"	can't get [\"b\"] from a non map type:",
		"	null",
	}, diff)
}
