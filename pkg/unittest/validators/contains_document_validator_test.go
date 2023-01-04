package validators_test

import (
	"testing"

	"github.com/lrills/helm-unittest/internal/common"
	. "github.com/lrills/helm-unittest/pkg/unittest/validators"
	"github.com/stretchr/testify/assert"
)

var docToTestContainsDocument1 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
`

var docToTestContainsDocument2 = `
apiVersion: v1
kind: Service
metadata:
  name: bar
  namespace: foo
`

func TestContainsDocumentValidatorWhenOk(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"v1",
		"bar",
		"foo",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsDocumentValidatorIndexWhenOk(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"v1",
		"bar",
		"foo",
	}
	pass, diff := validator.Validate(&ValidateContext{
		Index: 1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsDocumentValidatorNoNameWhenOk(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"v1",
		"",
		"foo",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsDocumentValidatorNoNamespaceWhenOk(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"v1",
		"foo",
		"",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsDocumentValidatorNoNameNamespaceWhenOk(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"v1",
		"",
		"",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.True(t, pass)
	assert.Equal(t, []string{}, diff)
}

func TestContainsDocumentValidatorWhenFailKind(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Deployment",
		"apps/v1",
		"foo",
		"bar",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Expected to contain document:",
		"\tKind = Deployment, apiVersion = apps/v1",
	}, diff)
}

func TestContainsDocumentValidatorWhenFailAPIVersion(t *testing.T) {
	validator := ContainsDocumentValidator{
		"Service",
		"apps/v1",
		"foo",
		"bar",
	}

	pass, diff := validator.Validate(&ValidateContext{
		Index: -1,
		Docs: []common.K8sManifest{makeManifest(docToTestContainsDocument1),
			makeManifest(docToTestContainsDocument2)},
	})

	assert.False(t, pass)
	assert.Equal(t, []string{
		"DocumentIndex:\t0",
		"Expected to contain document:",
		"\tKind = Service, apiVersion = apps/v1",
	}, diff)
}
