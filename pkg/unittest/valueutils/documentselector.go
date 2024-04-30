package valueutils

import (
	"errors"
	"github.com/helm-unittest/helm-unittest/internal/common"
	"reflect"
)

type DocumentSelector struct {
	MatchMany bool `yaml:"matchMany"`
	Path      string
	Value     interface{}
}

func (ds DocumentSelector) FilterDocuments(fileManifests []common.K8sManifest) ([]common.K8sManifest, error) {
	filteredManifests := []common.K8sManifest{}

	for _, doc := range fileManifests {
		var indexError error
		isMatchingSelector, indexError := ds.isMatchingSelector(doc)

		if indexError != nil {
			return filteredManifests, indexError
		} else if isMatchingSelector {
			if (!ds.MatchMany) && (len(filteredManifests) > 0) {
				return filteredManifests, errors.New("multiple indexes found")
			} else {
				filteredManifests = append(filteredManifests, doc)
			}
		}
	}

	if len(filteredManifests) > 0 {
		return filteredManifests, nil
	}

	return filteredManifests, errors.New("document not found")
}

func (ds DocumentSelector) isMatchingSelector(doc common.K8sManifest) (bool, error) {
	manifestValues, err := GetValueOfSetPath(doc, ds.Path)
	if err != nil {
		return false, err
	}

	for _, manifestValue := range manifestValues {
		if reflect.DeepEqual(ds.Value, manifestValue) {
			return true, nil
		}
	}

	return false, nil
}
