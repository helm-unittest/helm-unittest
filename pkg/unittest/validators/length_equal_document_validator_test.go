package validators_test

import (
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
	assert.Equal(t, []string{"\texpected result does not match"}, diff)
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
	assert.Equal(t, []string{"DocumentIndex:\t0", "Error:", "\tcount doesn't match as expected. expected: 1 actual: 2", "\texpected result does not match"}, diff)
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
	assert.Equal(t, []string{"\texpected result does not match"}, diff)
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
	assert.Equal(t, []string{"DocumentIndex:\t0", "Error:", "\tspec.tls count doesn't match as expected. actual: 2", "\texpected result does not match"}, diff)
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
