package snapshot

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// SnapshotFormat defines the format used to store snapshots
type SnapshotFormat string

const (
	// SnapshotFormatIndexed is the default format using numeric indexes (1, 2, 3...)
	SnapshotFormatIndexed SnapshotFormat = "indexed"
	// SnapshotFormatYAML stores snapshots as pure YAML with --- separators
	SnapshotFormatYAML SnapshotFormat = "yaml"
)

// CompareResult result return by Cache.Compare
type CompareResult struct {
	Passed         bool
	Test           string
	Index          uint
	NewSnapshot    string
	CachedSnapshot string
	Msg            string
	Err            error
}

// Cache manage snapshot caching
type Cache struct {
	Filepath      string
	Existed       bool
	IsUpdating    bool
	cached        map[string]map[uint]string
	current       map[string]map[uint]string
	// cachedYAML stores snapshots in pure YAML format (test name -> yaml string)
	cachedYAML map[string]string
	// currentYAML stores current snapshots in pure YAML format
	currentYAML   map[string]string
	updatedCount  uint
	insertedCount uint
	currentCount  uint
}

// RestoreFromFile restore cached snapshot from cache file
func (s *Cache) RestoreFromFile() error {
	content, err := os.ReadFile(s.Filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// First, unmarshal into a flexible structure to detect format per test
	var rawCache map[string]any
	if err := common.YmlUnmarshal(string(content), &rawCache); err != nil {
		return err
	}

	// Initialize maps
	s.cached = make(map[string]map[uint]string)
	s.cachedYAML = make(map[string]string)

	// Process each test entry and detect its format
	for testName, value := range rawCache {
		switch v := value.(type) {
		case string:
			// YAML format: value is a string containing YAML documents
			s.cachedYAML[testName] = v
		case map[string]any:
			// Indexed format: value is a map of index -> content
			indexedMap := make(map[uint]string)
			for key, val := range v {
				// Parse the key as uint
				var idx uint
				if _, err := fmt.Sscanf(key, "%d", &idx); err == nil {
					if strVal, ok := val.(string); ok {
						indexedMap[idx] = strVal
					}
				}
			}
			s.cached[testName] = indexedMap
		}
	}

	s.Existed = true
	return nil
}

func (s *Cache) getCached(test string, idx uint) (string, bool) {
	if cachedByTest, ok := s.cached[test]; ok {
		if cachedOfAssertion, ok := cachedByTest[idx]; ok {
			return cachedOfAssertion, true
		}
	}
	return "", false
}

// Compare content to cached last time, return CompareResult
func (s *Cache) Compare(test string, idx uint, content any, optFns ...func(options *CacheOptions) error) *CompareResult {
	var options CacheOptions
	var err error
	var msg string

	for _, optFn := range optFns {
		if err = optFn(&options); err != nil {
			options = CacheOptions{}
		}
	}

	s.currentCount++
	cached, existed := s.getCached(test, idx)
	if !existed {
		s.insertedCount++
	}

	match := true

	newSnapshot := common.TrustedMarshalYAML(content)

	if options.IsRegexEnabled() {
		if options.MatchRegexPattern != "" {
			match, err = valueutils.MatchesPattern(newSnapshot, options.MatchRegexPattern)
			if !match {
				msg = fmt.Sprintf(" pattern '%s' not found in snapshot", options.MatchRegexPattern)
			}
		}

		if options.NotMatchRegexPattern != "" && match {
			var noMatch bool
			noMatch, err = valueutils.MatchesPattern(newSnapshot, options.NotMatchRegexPattern)
			if noMatch {
				match = false
				msg = fmt.Sprintf(" pattern '%s' should not be in snapshot", options.NotMatchRegexPattern)
			}
		}
	} else {
		if existed && newSnapshot != cached {
			match = false
			s.updatedCount++
		}
	}

	var snapshotToSave string
	if s.IsUpdating || !existed {
		snapshotToSave = newSnapshot
	} else {
		snapshotToSave = cached
	}
	s.setNewSnapshot(test, idx, snapshotToSave)

	match = s.IsUpdating || match

	return &CompareResult{
		Passed:         match,
		Test:           test,
		Index:          idx,
		CachedSnapshot: cached,
		NewSnapshot:    newSnapshot,
		Msg:            msg,
		Err:            err,
	}
}

func (s *Cache) setNewSnapshot(test string, idx uint, snapshot string) {
	if s.current == nil {
		s.current = make(map[string]map[uint]string)
	}
	if newCacheOfTest, ok := s.current[test]; ok {
		newCacheOfTest[idx] = snapshot
	} else {
		s.current[test] = map[uint]string{idx: snapshot}
	}
}

// CompareYAML compares content in YAML format (all documents as a single string with --- separators)
// For YAML format, manifests are accumulated across multiple calls and compared at the end
func (s *Cache) CompareYAML(test string, contents []any, optFns ...func(options *CacheOptions) error) *CompareResult {
	var options CacheOptions
	var err error
	var msg string

	for _, optFn := range optFns {
		if err = optFn(&options); err != nil {
			options = CacheOptions{}
		}
	}

	// Build the new snapshot fragment as pure YAML with --- separators
	newSnapshotFragment := s.buildYAMLSnapshot(contents)

	// Accumulate content for this test
	s.setNewSnapshotYAML(test, newSnapshotFragment)

	// Get the accumulated content so far
	accumulated := s.currentYAML[test]

	// Check if this is a new test (for counting purposes)
	cached, existed := s.getCachedYAML(test)
	
	// Only count once per test (check if we've already counted)
	if _, alreadyCounted := s.current[test]; !alreadyCounted {
		s.currentCount++
		if !existed {
			s.insertedCount++
		}
		// Mark as counted using the indexed map (as a flag)
		if s.current == nil {
			s.current = make(map[string]map[uint]string)
		}
		s.current[test] = map[uint]string{0: "yaml-format-marker"}
	}

	match := true

	if options.IsRegexEnabled() {
		if options.MatchRegexPattern != "" {
			match, err = valueutils.MatchesPattern(accumulated, options.MatchRegexPattern)
			if !match {
				msg = fmt.Sprintf(" pattern '%s' not found in snapshot", options.MatchRegexPattern)
			}
		}

		if options.NotMatchRegexPattern != "" && match {
			var noMatch bool
			noMatch, err = valueutils.MatchesPattern(accumulated, options.NotMatchRegexPattern)
			if noMatch {
				match = false
				msg = fmt.Sprintf(" pattern '%s' should not be in snapshot", options.NotMatchRegexPattern)
			}
		}
	} else {
		// For non-regex mode, defer comparison until all content is accumulated
		// Just pass for now - final comparison happens in Changed() / StoreToFileIfNeeded()
		match = true
	}

	// When updating, always pass
	match = s.IsUpdating || match

	return &CompareResult{
		Passed:         match,
		Test:           test,
		Index:          0, // Not used in YAML format
		CachedSnapshot: cached,
		NewSnapshot:    accumulated,
		Msg:            msg,
		Err:            err,
	}
}

func (s *Cache) getCachedYAML(test string) (string, bool) {
	if s.cachedYAML == nil {
		return "", false
	}
	if cached, ok := s.cachedYAML[test]; ok {
		return cached, true
	}
	return "", false
}

func (s *Cache) setNewSnapshotYAML(test string, snapshot string) {
	if s.currentYAML == nil {
		s.currentYAML = make(map[string]string)
	}
	// Append to existing content instead of replacing
	// This allows accumulating manifests from multiple templates
	if existing, ok := s.currentYAML[test]; ok && existing != "" {
		// Append new content (remove leading --- since we'll join properly)
		newContent := strings.TrimPrefix(snapshot, "---\n")
		s.currentYAML[test] = existing + "---\n" + newContent
	} else {
		s.currentYAML[test] = snapshot
	}
}

// buildYAMLSnapshot builds a pure YAML string from multiple contents with --- separators
func (s *Cache) buildYAMLSnapshot(contents []any) string {
	if len(contents) == 0 {
		return ""
	}

	var parts []string
	for _, content := range contents {
		yamlStr := common.TrustedMarshalYAML(content)
		// Remove trailing newlines for consistent formatting
		yamlStr = strings.TrimSuffix(yamlStr, "\n")
		parts = append(parts, yamlStr)
	}

	// Join with YAML document separator
	return "---\n" + strings.Join(parts, "\n---\n") + "\n"
}

// Changed check if content have changed according to all Compare called
func (s *Cache) Changed() bool {
	if s.updatedCount > 0 || s.insertedCount > 0 {
		return true
	}

	// Check indexed format changes
	for test, cachedFiles := range s.cached {
		// Skip if this test is using YAML format (marked with yaml-format-marker)
		if idxMap, ok := s.current[test]; ok {
			if _, isYAMLFormat := idxMap[0]; isYAMLFormat && idxMap[0] == "yaml-format-marker" {
				continue
			}
		}
		if _, ok := s.current[test]; !ok {
			return true
		}
		for idx := range cachedFiles {
			if _, ok := s.current[test][idx]; !ok {
				return true
			}
		}
	}

	// Check YAML format changes - compare accumulated content with cached
	for test, cachedContent := range s.cachedYAML {
		currentContent, ok := s.currentYAML[test]
		if !ok {
			return true
		}
		if currentContent != cachedContent {
			return true
		}
	}

	// Check for new YAML format tests that weren't cached before
	for test := range s.currentYAML {
		if _, ok := s.cachedYAML[test]; !ok {
			return true
		}
	}

	return false
}

// StoreToFileIfNeeded store current cache to file if snapshot content changed
func (s *Cache) StoreToFileIfNeeded() (bool, error) {
	if !s.Changed() {
		return false, nil
	}

	if s.IsUpdating || s.insertedCount > 0 || s.VanishedCount() > 0 {
		// Merge both indexed and YAML format snapshots
		mergedSnapshots := s.mergeSnapshots()

		byteBuffer := new(bytes.Buffer)
		yamlEncoder := common.YamlNewEncoder(byteBuffer)
		yamlEncoder.SetIndent(common.YAMLINDENTION)
		if err := yamlEncoder.Encode(mergedSnapshots); err != nil {
			return false, err
		}

		if err := os.WriteFile(s.Filepath, byteBuffer.Bytes(), 0644); err != nil {
			return false, err
		}

		s.Existed = true
		return true, nil
	}

	return false, nil
}

// mergeSnapshots merges indexed and YAML format snapshots into a single map
// YAML format snapshots are stored directly as strings
// Indexed format snapshots are stored as map[uint]string
func (s *Cache) mergeSnapshots() map[string]any {
	merged := make(map[string]any)

	// Add indexed format snapshots (skip yaml-format-marker entries)
	for test, idxMap := range s.current {
		// Skip YAML format marker entries
		if len(idxMap) == 1 {
			if marker, ok := idxMap[0]; ok && marker == "yaml-format-marker" {
				continue
			}
		}
		merged[test] = idxMap
	}

	// Add YAML format snapshots (will overwrite if same test name exists)
	for test, yamlContent := range s.currentYAML {
		merged[test] = yamlContent
	}

	return merged
}

// UpdatedCount return snapshot count that was cached before and updated current time
func (s *Cache) UpdatedCount() uint {
	return s.updatedCount
}

// InsertedCount return snapshot count that was newly inserted current time
func (s *Cache) InsertedCount() uint {
	return s.insertedCount
}

// CurrentCount return total snapshot count of current time
func (s *Cache) CurrentCount() uint {
	return s.currentCount
}

// FailedCount return snapshot count that was failed when Compare
func (s *Cache) FailedCount() uint {
	if s.IsUpdating {
		return 0
	}
	return s.updatedCount
}

// VanishedCount return snapshot count that was cached last time but not exists this time
func (s *Cache) VanishedCount() uint {
	var count uint

	// Check indexed format
	for test, cachedFiles := range s.cached {
		for idx := range cachedFiles {
			if newTestCache, ok := s.current[test]; ok {
				if _, ok := newTestCache[idx]; ok {
					continue
				}
			}
			count++
		}
	}

	// Check YAML format
	for test := range s.cachedYAML {
		if _, ok := s.currentYAML[test]; !ok {
			count++
		}
	}

	return count
}

// CacheOptionsFunc is a type alias for CacheOptions functional option
type CacheOptionsFunc func(*CacheOptions) error

type CacheOptions struct {
	MatchRegexPattern    string
	NotMatchRegexPattern string
}

func WithMatchRegexPattern(pattern string) CacheOptionsFunc {
	return func(c *CacheOptions) error {
		c.MatchRegexPattern = strings.TrimSpace(pattern)
		return nil
	}
}

func WithNotMatchRegexPattern(pattern string) CacheOptionsFunc {
	return func(c *CacheOptions) error {
		c.NotMatchRegexPattern = strings.TrimSpace(pattern)
		return nil
	}
}

func (c *CacheOptions) IsRegexEnabled() bool {
	return c.NotMatchRegexPattern != "" || c.MatchRegexPattern != ""
}
