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
