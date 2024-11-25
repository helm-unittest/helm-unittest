package validators_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/validators"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var docToTestEqualRaw = `
raw: This is a NOTES.txt document.
`

func TestEqualRawValidatorWhenOk(t *testing.T) {
	manifest := makeManifest(docToTestEqualRaw)
	validator := EqualRawValidator{"This is a NOTES.txt document."}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestEqualRawValidatorWhenFail(t *testing.T) {
	manifest := makeManifest(docToTestEqualRaw)

	log.SetLevel(log.DebugLevel)

	validator := EqualRawValidator{"Invalid text."}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected to equal:",
		"	Invalid text.",
		"Actual:",
		"	This is a NOTES.txt document.",
		"Diff:",
		"	--- Expected",
		"	+++ Actual",
		"	@@ -1,2 +1,2 @@",
		"	-Invalid text.",
		"	+This is a NOTES.txt document.",
	}, diff)
}

func TestEqualRawValidatorWhenNegativeAndFail(t *testing.T) {
	manifest := makeManifest(docToTestEqualRaw)

	v := EqualRawValidator{"This is a NOTES.txt document."}
	pass, diff := v.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to equal:",
		"	This is a NOTES.txt document.",
	}, diff)
}

func TestEqualRawValidatorWhenNegativeAndFailFast(t *testing.T) {
	manifest := makeManifest(docToTestEqualRaw)

	v := EqualRawValidator{"This is a NOTES.txt document."}
	pass, diff := v.Validate(&ValidateContext{
		FailFast: true,
		Docs:     []common.K8sManifest{manifest, manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected NOT to equal:",
		"	This is a NOTES.txt document.",
	}, diff)
}

func TestEqualRawValidatorWhenNoManifestFail(t *testing.T) {
	validator := EqualRawValidator{"This is a NOTES.txt document."}

	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"Expected to equal:",
		"\tThis is a NOTES.txt document.",
		"Actual:",
		"\tno manifest found",
		"Diff:", "\t--- Expected",
		"\t+++ Actual",
		"\t@@ -1,2 +1,2 @@",
		"\t-This is a NOTES.txt document.",
		"\t+no manifest found"}, diff)
}

func TestEqualRawValidatorWhenNoManifestNegativeOk(t *testing.T) {
	validator := EqualRawValidator{"This is a NOTES.txt document."}

	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{},
		Negative: true,
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
