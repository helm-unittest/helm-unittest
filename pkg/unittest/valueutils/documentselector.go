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
func (ds DocumentSelector) FindDocumentsIndex(manifests map[string][]common.K8sManifest) (int, error) {
	indexFound := false
	actualIndex := -1
	for _, fileManifests := range manifests {
		for idx, doc := range fileManifests {
			var indexError error
			actualIndex, indexFound, indexError = ds.findDocumentIndex(doc, indexFound, actualIndex, idx)
			if indexError != nil {
				return actualIndex, indexError
			}
		}
	}

	if indexFound {
		return actualIndex, nil
	}

	return actualIndex, errors.New("document not found")
}

func (ds DocumentSelector) findDocumentIndex(doc common.K8sManifest, indexFound bool, currentIndex, idx int) (int, bool, error) {
	foundIndex := currentIndex
	manifestValues, err := GetValueOfSetPath(doc, ds.Path)
	if err != nil {
		return foundIndex, false, err
	}

	for _, manifestValue := range manifestValues {
		if reflect.DeepEqual(ds.Value, manifestValue) {
			if !indexFound {
				indexFound = true
				foundIndex = idx
			} else {
				return foundIndex, indexFound, errors.New("multiple indexes found")
			}
		}
	}

	return foundIndex, indexFound, nil
}
