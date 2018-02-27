package snapshot

import (
	"io/ioutil"
	"os"

	"github.com/lrills/helm-unittest/unittest/common"
	yaml "gopkg.in/yaml.v2"
)

// CompareResult result return by Cache.Compare
type CompareResult struct {
	Passed bool
	Test   string
	Index  int
	New    string
	Cached string
}

// Cache manage snapshot caching
type Cache struct {
	Filepath   string
	Existed    bool
	IsUpdating bool
	cached     map[string]map[int]string
	new        map[string]map[int]string
	updated    bool
}

// RestoreFromFile restore cached snapshot from cache file
func (s *Cache) RestoreFromFile() error {
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

func (s *Cache) getCached(test string, idx int) (string, bool) {
	if cachedByTest, ok := s.cached[test]; ok {
		if cachedOfAssertion, ok := cachedByTest[idx]; ok {
			return cachedOfAssertion, true
		}
	}
	return "", false
}

// Compare compare content to cached last time, return CompareResult
func (s *Cache) Compare(test string, idx int, content interface{}) *CompareResult {
	newSnapshot := s.saveNewCache(test, idx, content)
	match, cachedSnapshot := s.compareToCached(test, idx, newSnapshot)
	return &CompareResult{
		Passed: s.IsUpdating || match,
		Test:   test,
		Index:  idx,
		Cached: cachedSnapshot,
		New:    newSnapshot,
	}
}

func (s *Cache) compareToCached(test string, idx int, snapshot string) (bool, string) {
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

func (s *Cache) saveNewCache(test string, idx int, content interface{}) string {
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

// Changed check if content have changed according to all Compare called
func (s *Cache) Changed() bool {
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

// StoreToFile store new cache to file
func (s *Cache) StoreToFile() error {
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
