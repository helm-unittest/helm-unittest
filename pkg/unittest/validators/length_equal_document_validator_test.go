package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
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

	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: 1,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}
func TestLengthEqualDocumentsValidatorOk_Single2(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual2)

	validator := LengthEqualDocumentsValidator{
		Path:  "spec.tls",
		Count: 2,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
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

func TestLengthEqualDocumentsValidatorFail_Multi(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls", "spec.rules"},
	}
	pass, _ := validator.Validate(&ValidateContext{
		Docs: []common.K8sManifest{manifest},
	})

	assert.False(t, pass)
	//assert.Equal(t, []string{}, diff)
}
func TestLengthEqualDocumentsValidatorWhenBadConfig(t *testing.T) {
	manifest := makeManifest(testDocLengthEqual3_Fail)

	validator := LengthEqualDocumentsValidator{
		Paths: []string{"spec.tls"},
		Path:  "spec.tls",
		Count: 2,
	}
	pass, diff := validator.Validate(&ValidateContext{
		Docs:     []common.K8sManifest{manifest},
		Negative: true,
	})

	assert.False(t, pass)
	assert.Equal(t, []string{"Error:", "\t'paths' couldn't be used with 'path'"}, diff)
}
