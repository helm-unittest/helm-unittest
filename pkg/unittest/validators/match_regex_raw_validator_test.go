package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestMatchRegexRaw = `
raw: |
  This is a NOTES.txt document.
`

func TestMatchRegexRawValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexRaw)

	validator := MatchRegexRawValidator{"^This"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexRawValidatorWhenNegativeAndOk(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexRaw)

	validator := MatchRegexRawValidator{"^foo"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})
	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestMatchRegexRawValidatorWhenRegexCompileFail(t *testing.T) {
	manifest := common.K8sManifest{"raw": ""}

	validator := MatchRegexRawValidator{"+"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	error parsing regexp: missing argument to repetition operator: `+`",
	}, diff)
}

func TestMatchRegexRawValidatorWhenMatchFail(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexRaw)

	log.SetLevel(log.DebugLevel)

	validator := MatchRegexRawValidator{"^foo"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})
	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected to match:",
		"	^foo",
		"Actual:",
		"	This is a NOTES.txt document.",
	}, diff)
}

func TestMatchRegexRawValidatorWhenNegativeAndMatchFail(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegexRaw)

	validator := MatchRegexRawValidator{"^This"}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to match:",
		"	^This",
		"Actual:",
		"	This is a NOTES.txt document.",
	}, diff)
}

func TestMatchRegexRawValidatorWhenNoPattern(t *testing.T) {
	manifest := makeManifest(docToTestMatchRegex)

	validator := MatchRegexRawValidator{""}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Error:",
		"	expected field 'pattern' to be filled",
	}, diff)
}
