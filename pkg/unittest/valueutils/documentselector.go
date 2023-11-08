package valueutils

import (
	"errors"
	"reflect"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

type DocumentSelector struct {
	Path  string
	Value interface{}
}

// FindDocumentIndex, find the index of a document, based on a jsonyamlpath and value
func FindDocumentIndex(manifests map[string][]common.K8sManifest, selector DocumentSelector) (int, error) {
	for _, fileManifests := range manifests {
		for idx, doc := range fileManifests {
			manifestValues, err := GetValueOfSetPath(doc, selector.Path)
			if err != nil {
				continue
			}

			for _, manifestValue := range manifestValues {
				if reflect.DeepEqual(selector.Value, manifestValue) {
					return idx, nil
				}
			}
		}
	}

	return -1, errors.New("document not found")
}
