package valueutils_test

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

var docToTestIndex0 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
  service: internal
`
var docToTestIndex1 = `
apiVersion: v1
kind: Service
metadata:
  name: foo
  namespace: bar
`

func createMultiManifest() []common.K8sManifest {
	manifest1 := common.K8sManifest{}
	yaml.Unmarshal([]byte(docToTestIndex0), &manifest1)
	manifest2 := common.K8sManifest{}
	yaml.Unmarshal([]byte(docToTestIndex1), &manifest2)

	return []common.K8sManifest{manifest1, manifest2}
}

func TestFindDocumentsIndexSinglePathOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := []common.K8sManifest{createMultiManifest()[0]}

	selector := DocumentSelector{
		Path:  "metadata.service",
		Value: "internal",
	}

	actualManifests, err := selector.FilterDocuments(createMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexObjectValueOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := []common.K8sManifest{createMultiManifest()[1]}

	selector := DocumentSelector{
		Path: "metadata",
		Value: map[string]interface{}{
			"name":      "foo",
			"namespace": "bar",
		},
	}

	actualManifests, err := selector.FilterDocuments(createMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexMultiIndexNOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := []common.K8sManifest{createMultiManifest()[0]}

	selector := DocumentSelector{
		Path:  "metadata.name",
		Value: "foo",
	}

	actualManifests, err := selector.FilterDocuments(createMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "multiple indexes found")
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndicesMultiAllowedIndexOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := createMultiManifest()

	selector := DocumentSelector{
		Path:      "metadata.name",
		Value:     "foo",
		MatchMany: true,
	}

	actualManifests, err := selector.FilterDocuments(createMultiManifest())

	a.Nil(err)
	a.Equal(expectedManifests, actualManifests)
}

func TestFindDocumentIndexNoDocumentNOk(t *testing.T) {
	a := assert.New(t)
	expectedManifests := []common.K8sManifest{}

	selector := DocumentSelector{
		Path:  "meta.data",
		Value: "bar",
	}

	actualManifests, err := selector.FilterDocuments(createMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "document not found")
	a.Equal(expectedManifests, actualManifests)
}
