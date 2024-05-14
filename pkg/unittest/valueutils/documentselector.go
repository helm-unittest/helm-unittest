package valueutils

import (
	"errors"
	"reflect"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

type DocumentSelector struct {
	SkipEmptyTemplates bool `yaml:"skipEmptyTemplates"`
	MatchMany          bool `yaml:"matchMany"`
	Path               string
	Value              interface{}
}

func (ds DocumentSelector) SelectDocuments(documentsByTemplate map[string][]common.K8sManifest) (map[string][]common.K8sManifest, error) {

	matchingDocuments := map[string][]common.K8sManifest{}
	matchingDocumentsCount := 0

	for template, manifests := range documentsByTemplate {
		filteredManifests, err := ds.selectDocuments(manifests)

		filteredManifestsCount := len(filteredManifests)
		matchingDocumentsCount += filteredManifestsCount

		if err != nil {
			return map[string][]common.K8sManifest{}, err
		}

		if !ds.MatchMany && matchingDocumentsCount > 1 {
			return map[string][]common.K8sManifest{}, errors.New("multiple indexes found")
		}

		if filteredManifestsCount > 0 || !ds.SkipEmptyTemplates {
			matchingDocuments[template] = filteredManifests
		}
	}

	return matchingDocuments, nil
}

func (ds DocumentSelector) selectDocuments(docs []common.K8sManifest) ([]common.K8sManifest, error) {
	selectedDocs := []common.K8sManifest{}

	for _, doc := range docs {
		var indexError error
		isMatchingSelector, indexError := ds.isMatchingSelector(doc)

		if indexError != nil {
			return selectedDocs, indexError
		} else if isMatchingSelector {
			if (!ds.MatchMany) && (len(selectedDocs) > 0) {
				return selectedDocs, errors.New("multiple indexes found")
			} else {
				selectedDocs = append(selectedDocs, doc)
			}
		}
	}

	if ds.SkipEmptyTemplates || len(selectedDocs) > 0 {
		return selectedDocs, nil
	}

	return selectedDocs, errors.New("document not found")
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
