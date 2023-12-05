package valueutils_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func createMultiManifest() map[string][]common.K8sManifest {
	manifest1 := common.K8sManifest{}
	yaml.Unmarshal([]byte(docToTestIndex0), &manifest1)
	manifest2 := common.K8sManifest{}
	yaml.Unmarshal([]byte(docToTestIndex1), &manifest2)

	manifestArray := []common.K8sManifest{manifest1, manifest2}
	multiManifest := map[string][]common.K8sManifest{
		"service.yaml": manifestArray,
	}

	return multiManifest
}

func TestFindDocumentsIndexSinglePathOk(t *testing.T) {
	a := assert.New(t)
	expectedIndex := 0

	selector := DocumentSelector{
		Path:  "metadata.service",
		Value: "internal",
	}

	actualIndex, err := selector.FindDocumentsIndex(createMultiManifest())

	a.Nil(err)
	a.Equal(expectedIndex, actualIndex)
}

func TestFindDocumentIndexObjectValueOk(t *testing.T) {
	a := assert.New(t)
	expectedIndex := 1

	selector := DocumentSelector{
		Path: "metadata",
		Value: map[string]interface{}{
			"name":      "foo",
			"namespace": "bar",
		},
	}

	actualIndex, err := selector.FindDocumentsIndex(createMultiManifest())

	a.Nil(err)
	a.Equal(expectedIndex, actualIndex)
}

func TestFindDocumentIndexMultiIndexNOk(t *testing.T) {
	a := assert.New(t)
	expectedIndex := 0

	selector := DocumentSelector{
		Path:  "metadata.name",
		Value: "foo",
	}

	actualIndex, err := selector.FindDocumentsIndex(createMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "multiple indexes found")
	a.Equal(expectedIndex, actualIndex)
}

func TestFindDocumentIndexNoDocumentNOk(t *testing.T) {
	a := assert.New(t)
	expectedIndex := -1

	selector := DocumentSelector{
		Path:  "meta.data",
		Value: "bar",
	}

	actualIndex, err := selector.FindDocumentsIndex(createMultiManifest())

	a.NotNil(err)
	a.EqualError(err, "document not found")
	a.Equal(expectedIndex, actualIndex)
}
