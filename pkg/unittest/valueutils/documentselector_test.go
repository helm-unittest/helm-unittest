package valueutils_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
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
	return common.TrustedUnmarshalYAML(manifest)
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

func TestNewSafeDocumentSelector_Success(t *testing.T) {
	tests := []struct {
		name             string
		input            map[string]interface{}
		expectedSelector *DocumentSelector
	}{
		{
			name: "all fields set",
			input: map[string]interface{}{
				"path":               "metadata.name",
				"value":              "foo",
				"matchMany":          true,
				"skipEmptyTemplates": true,
			},
			expectedSelector: &DocumentSelector{
				Path:               "metadata.name",
				Value:              "foo",
				MatchMany:          true,
				SkipEmptyTemplates: true,
			},
		},
		{
			name: "only required fields set",
			input: map[string]interface{}{
				"path":  "metadata.name",
				"value": "foo",
			},
			expectedSelector: &DocumentSelector{
				Path:               "metadata.name",
				Value:              "foo",
				MatchMany:          false,
				SkipEmptyTemplates: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			actualSelector, err := NewDocumentSelector(tt.input)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedSelector, actualSelector)
		})
	}
}

func TestNewDocumentSelectorMissingPath(t *testing.T) {
	input := map[string]interface{}{
		"value": "foo",
	}

	selector, err := NewDocumentSelector(input)

	assert.NotNil(t, err)
	assert.Nil(t, selector)
}
