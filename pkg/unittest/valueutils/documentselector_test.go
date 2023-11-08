package valueutils_test

import (
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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

func TestFindDocumentIndexSinglePathOk(t *testing.T) {
	a := assert.New(t)
	expectedIndex := 0

	selector := DocumentSelector{
		Path:  "metadata.service",
		Value: "internal",
	}

	actualIndex, err := FindDocumentIndex(createMultiManifest(), selector)

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

	actualIndex, err := FindDocumentIndex(createMultiManifest(), selector)

	a.Nil(err)
	a.Equal(expectedIndex, actualIndex)
}
