package snapshot

import (
	"fmt"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/valueutils"
)

// recordedSnapshot is one snapshot a test case wants stored in the cache.
type recordedSnapshot struct {
	test     string
	idx      uint
	snapshot string
}

// Recorder buffers the snapshot comparisons of a single test case. Comparing only
// reads the snapshots restored from file, so the recorders of different test cases
// may compare concurrently; Cache.Commit applies a recorder to its cache afterwards.
type Recorder struct {
	cache    *Cache
	entries  []recordedSnapshot
	current  uint
	inserted uint
	updated  uint
}

// NewRecorder returns an empty recorder comparing against this cache.
func (s *Cache) NewRecorder() *Recorder {
	return &Recorder{cache: s}
}

// Compare content to cached last time, return CompareResult. The snapshot to store
// and the counts are kept in the recorder until it is committed.
func (r *Recorder) Compare(test string, idx uint, content any, optFns ...func(options *CacheOptions) error) *CompareResult {
	var options CacheOptions
	var err error
	var msg string

	for _, optFn := range optFns {
		if err = optFn(&options); err != nil {
			options = CacheOptions{}
		}
	}

	r.current++
	cached, existed := r.cache.getCached(test, idx)
	if !existed {
		r.inserted++
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
			r.updated++
		}
	}

	var snapshotToSave string
	if r.cache.IsUpdating || !existed {
		snapshotToSave = newSnapshot
	} else {
		snapshotToSave = cached
	}
	r.entries = append(r.entries, recordedSnapshot{test: test, idx: idx, snapshot: snapshotToSave})

	match = r.cache.IsUpdating || match

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

// Commit applies the snapshots and counts buffered by the recorder to the cache.
// Commits must not run concurrently with each other.
func (s *Cache) Commit(r *Recorder) {
	s.currentCount += r.current
	s.insertedCount += r.inserted
	s.updatedCount += r.updated
	for _, entry := range r.entries {
		s.setNewSnapshot(entry.test, entry.idx, entry.snapshot)
	}
}
