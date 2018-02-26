package snapshot

import (
	"io/ioutil"
	"os"

	"github.com/lrills/helm-unittest/unittest/common"
	yaml "gopkg.in/yaml.v2"
)

type SnapshotCompareResult struct {
	Test    string
	Index   int
	Matched bool
	New     string
	Cached  string
}

type SnapshotCache struct {
	Filepath string
	Existed  bool
	cached   map[string]map[int]string
	new      map[string]map[int]string
	updated  bool
}

func (s *SnapshotCache) RestoreFromFile() error {
	content, err := ioutil.ReadFile(s.Filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := yaml.Unmarshal(content, &s.cached); err != nil {
		return err
	}
	s.Existed = true
	return nil
}

func (s *SnapshotCache) getCached(test string, idx int) (string, bool) {
	if cachedByTest, ok := s.cached[test]; ok {
		if cachedOfAssertion, ok := cachedByTest[idx]; ok {
			return cachedOfAssertion, true
		}
	}
	return "", false
}

func (s *SnapshotCache) Compare(test string, idx int, content interface{}) *SnapshotCompareResult {
	newSnapshot := s.saveNewCache(test, idx, content)
	match, cachedSnapshot := s.compareToCached(test, idx, newSnapshot)
	return &SnapshotCompareResult{
		Test:    test,
		Index:   idx,
		Matched: match,
		Cached:  cachedSnapshot,
		New:     newSnapshot,
	}
}

func (s *SnapshotCache) compareToCached(test string, idx int, snapshot string) (bool, string) {
	cached, ok := s.getCached(test, idx)
	if !ok {
		s.updated = true
		return false, ""
	}

	if snapshot != cached {
		s.updated = true
		return false, cached
	}
	return true, cached
}

func (s *SnapshotCache) saveNewCache(test string, idx int, content interface{}) string {
	snapshot := common.TrustedMarshalYAML(content)
	if s.new == nil {
		s.new = make(map[string]map[int]string)
	}
	if newCacheOfTest, ok := s.new[test]; ok {
		newCacheOfTest[idx] = snapshot
	} else {
		s.new[test] = map[int]string{idx: snapshot}
	}
	return snapshot
}

func (s *SnapshotCache) ShouldUpdate() bool {
	if s.updated {
		return true
	}

	for test, cachedFiles := range s.cached {
		if _, ok := s.new[test]; !ok {
			return true
		}
		for idx := range cachedFiles {
			if _, ok := s.new[test][idx]; !ok {
				return true
			}
		}
	}
	return false
}

func (s *SnapshotCache) StoreToFile() error {
	cacheData, err := yaml.Marshal(s.new)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(s.Filepath, cacheData, 0644); err != nil {
		return err
	}
	s.Existed = true
	return nil
}
