package valueutils_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var firstTemplateDocToTestIndex0 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
  service: internal
`
var firstTemplateDocToTestIndex1 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
`

var secondTemplateDocToTestIndex0 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: foo
`

var secondTemplateDocToTestIndex1 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: foo
`

func createSingleTemplateMultiManifest() map[string][]common.K8sManifest {
	return map[string][]common.K8sManifest{
		"firstTemplate": {
			parseManifest(firstTemplateDocToTestIndex0), parseManifest(firstTemplateDocToTestIndex1),
		},
	}
}

func createMultiTemplateMultiManifest() map[string][]common.K8sManifest {
	return map[string][]common.K8sManifest{
		"firstTemplate": {
			parseManifest(firstTemplateDocToTestIndex0), parseManifest(firstTemplateDocToTestIndex1),
		},
		"secondTemplate": {
			parseManifest(secondTemplateDocToTestIndex0), parseManifest(secondTemplateDocToTestIndex1),
		},
	}
}

func parseManifest(manifest string) common.K8sManifest {
	parsedManifest := common.K8sManifest{}
	yaml.Unmarshal([]byte(manifest), &parsedManifest)

	return parsedManifest
}

func TestFindDocumentsIndexSinglePathOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{"firstTemplate": {parseManifest(firstTemplateDocToTestIndex0)}}

	selector := DocumentSelector{
		Path:  "metadata.service",
		Value: "internal",
	}

	actualManifests, err := selector.SelectDocuments(createSingleTemplateMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexObjectValueOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{"firstTemplate": {parseManifest(firstTemplateDocToTestIndex1)}}

	selector := DocumentSelector{
		Path: "metadata",
		Value: map[string]interface{}{
			"name":      "foo",
			"namespace": "bar",
		},
	}

	actualManifests, err := selector.SelectDocuments(createSingleTemplateMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexMultiIndexNOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{}

	selector := DocumentSelector{
		Path:  "metadata.name",
		Value: "foo",
	}

	actualManifests, err := selector.SelectDocuments(createSingleTemplateMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "multiple indexes found")
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndicesMultiAllowedIndexOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := createSingleTemplateMultiManifest()

	selector := DocumentSelector{
		Path:      "metadata.name",
		Value:     "foo",
		MatchMany: true,
	}

	actualManifests, err := selector.SelectDocuments(createSingleTemplateMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexNoDocumentNOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{}

	selector := DocumentSelector{
		Path:  "meta.data",
		Value: "bar",
	}

	actualManifests, err := selector.SelectDocuments(createSingleTemplateMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "document not found")
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndicesMatchManyAndSkipEmptyTemplatesOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{
		"secondTemplate": {parseManifest(secondTemplateDocToTestIndex0), parseManifest(secondTemplateDocToTestIndex1)},
	}

	selector := DocumentSelector{
		Path:               "metadata.namespace",
		Value:              "foo",
		MatchMany:          true,
		SkipEmptyTemplates: true,
	}

	actualManifests, err := selector.SelectDocuments(createMultiTemplateMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndicesMatchManyAndDontSkipEmptyTemplatesNOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := map[string][]common.K8sManifest{}

	selector := DocumentSelector{
		Path:               "metadata.namespace",
		Value:              "foo",
		MatchMany:          true,
		SkipEmptyTemplates: false,
	}

	actualManifests, err := selector.SelectDocuments(createMultiTemplateMultiManifest())

	a.EqualError(err, "document not found")
	a.Equal(expectedManifests, actualManifests)
}
