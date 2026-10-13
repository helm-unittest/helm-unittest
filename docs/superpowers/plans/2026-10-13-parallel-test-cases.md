# Parallel Test Cases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `--parallel-level=suite|test` so that, with `--parallel`, individual test cases of all suites run concurrently in one shared worker pool while output, counts and `.snap` files stay identical to a sequential run.

**Architecture:** Snapshot comparisons are split into a read-only compare that buffers into a per-case `snapshot.Recorder` and a `Cache.Commit` applied by the suite in index order, so the cache needs no lock and failfast can drop cases cleanly. `TestSuite.runV4TestJobs` gains a parallel path that runs each case on a slot of a pool owned by the runner; only cases hold slots, coordinators never do, so the pool cannot deadlock. The runner dispatches `--parallel-level=test` to the existing parallel group runner with an unbounded group limit plus the case pool.

**Tech Stack:** Go 1.26, `sync`, `sync/atomic`, testify, cobra, logrus, Helm v4 chart libraries.

**Spec:** `docs/superpowers/specs/2026-10-13-parallel-test-cases-design.md`

## Global Constraints

- `--parallel-level` values: exactly `suite` (default) and `test`; anything else is a CLI error before any chart runs.
- `--parallel-level` only has effect with `--parallel`; `test` without `--parallel` logs a warning and runs sequentially.
- `--debugPlugin` (non-empty `RenderPath`) forces sequential execution regardless of both flags.
- Behaviour and output of `--parallel` with the default `suite` level must not change.
- At `test` level, concurrently running test cases never exceed `effectiveWorkers(totalCases)` (`--max-workers`, `0` = CPU count).
- No new module dependencies.
- All tests pass under `CGO_ENABLED=1 go test ./... -race` (`make unittest`).
- CHANGELOG entries go in the unreleased `1.3.0` section; do not bump `plugin.yaml`.
- Commit messages end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. A suite whose cases are **all skipped by `skip.reason`** at `test` level must report `Skipped=true`, `Passed=true` and record no snapshots, exactly as sequential. Pinned in Task 4 (`TestRunV4TestJobsParallelAllSkipped`).
2. **Two cases with the same `it:` name** writing snapshots under `-u` must produce the same `.snap` bytes on every run and equal sequential (last one wins). Pinned in Task 4 (`TestRunV4TestJobsParallelDuplicateNamesMatchSequential`) and Task 5 (`TestV4RunnerTestLevelUpdateSnapshotMatchesSequential`).
3. **`--failfast` with cases after the failure already finished** must leave no trace: not printed, not counted, not stored in `.snap`. Pinned in Task 4 (`TestRunV4TestJobsParallelFailfastMatchesSequential`) and Task 5 (`TestV4RunnerTestLevelFailfastMatchesSequential`).
4. **`--coverage` together with `--parallel-level=test`** must print the same coverage table as sequential. Pinned in Task 5 (`TestV4RunnerTestLevelCoverageMatchesSequential`).
5. **`--parallel-level=test` with `--debugPlugin`** must not run cases on the pool (debug output files would collide). Pinned in Task 5 (`TestRunV4TestLevelIgnoredWithDebugRender`).

## File Structure

| File | Change | Responsibility |
|---|---|---|
| `pkg/unittest/snapshot/recorder.go` | Create | `Recorder` (per-case buffered compare) and `Cache.NewRecorder` / `Cache.Commit` |
| `pkg/unittest/snapshot/cache.go` | Modify | `Cache.Compare` delegates to a recorder and commits immediately |
| `pkg/unittest/snapshot/cache_test.go` | Modify | Recorder tests |
| `pkg/unittest/models.go` | Modify | `TestConfig.snapshotRecorder`, `WithSnapshotRecorder` option |
| `pkg/unittest/test_job.go` | Modify | `orderedSnapshotComparer` targets a recorder or the cache |
| `pkg/unittest/test_job_test.go` | Modify | Recorder wiring test |
| `test/data/v3/parallel-testcases/**` | Create | Fixture chart for all new tests |
| `pkg/unittest/test_runner.go` | Modify | Failfast store fix, `ParallelLevel`, case pool dispatch, failfast print cutoff |
| `pkg/unittest/test_suite.go` | Modify | `jobAccumulator`, `runV4TestJob`, parallel case path |
| `pkg/unittest/test_suite_parallel_internal_test.go` | Create | Suite-level parallel tests |
| `pkg/unittest/test_runner_test.go` | Modify | Runner-level parity / failfast / coverage tests |
| `pkg/unittest/test_runner_internal_test.go` | Modify | Pool-size and debug-render tests |
| `cmd/helm-unittest/helm_unittest.go` | Modify | `--parallel-level` flag, validation, warning |
| `cmd/helm-unittest/helm_unittest_test.go` | Modify | Flag tests |
| `README.md`, `CHANGELOG.md` | Modify | Docs |

---

### Task 1: Snapshot recorder

**Files:**
- Create: `pkg/unittest/snapshot/recorder.go`
- Modify: `pkg/unittest/snapshot/cache.go` (the `Compare` method, lines ~92-155, and the import block)
- Test: `pkg/unittest/snapshot/cache_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces:
  - `func (s *Cache) NewRecorder() *Recorder`
  - `func (r *Recorder) Compare(test string, idx uint, content any, optFns ...func(options *CacheOptions) error) *CompareResult`
  - `func (s *Cache) Commit(r *Recorder)` — must not be called concurrently.
  - `Cache.Compare` keeps its exact signature and behaviour.

- [ ] **Step 1: Write the failing tests**

Append to `pkg/unittest/snapshot/cache_test.go` (add `"fmt"` and `"sync"` to its imports):

```go
func TestRecorderDoesNotTouchCacheUntilCommit(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	a.Nil(cache.RestoreFromFile())

	recorder := cache.NewRecorder()
	result := recorder.Compare(cache_before, 1, content1)

	a.Equal(createCacheResult(1, true, snapshot1, snapshot1), result)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Commit(recorder)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)
}

func TestRecorderMatchesCompare(t *testing.T) {
	cases := []struct {
		name     string
		existed  bool
		updating bool
		test     string
		content  any
		opts     []func(*CacheOptions) error
	}{
		{name: "insert", existed: false, test: "new test", content: content1},
		{name: "match", existed: true, test: cache_before, content: content1},
		{name: "mismatch", existed: true, test: cache_before, content: contentNew},
		{name: "update mismatch", existed: true, updating: true, test: cache_before, content: contentNew},
		{name: "regex", existed: true, test: cache_before, content: content1,
			opts: []func(*CacheOptions) error{WithMatchRegexPattern("a:")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			direct := createCache(a, tc.existed)
			direct.IsUpdating = tc.updating
			a.Nil(direct.RestoreFromFile())
			buffered := createCache(a, tc.existed)
			buffered.IsUpdating = tc.updating
			a.Nil(buffered.RestoreFromFile())

			directResult := direct.Compare(tc.test, 1, tc.content, tc.opts...)
			recorder := buffered.NewRecorder()
			bufferedResult := recorder.Compare(tc.test, 1, tc.content, tc.opts...)
			buffered.Commit(recorder)

			a.Equal(directResult, bufferedResult)
			a.Equal(direct.CurrentCount(), buffered.CurrentCount())
			a.Equal(direct.InsertedCount(), buffered.InsertedCount())
			a.Equal(direct.UpdatedCount(), buffered.UpdatedCount())
			a.Equal(direct.FailedCount(), buffered.FailedCount())
			a.Equal(direct.VanishedCount(), buffered.VanishedCount())

			directStored, directErr := direct.StoreToFileIfNeeded()
			bufferedStored, bufferedErr := buffered.StoreToFileIfNeeded()
			a.Nil(directErr)
			a.Nil(bufferedErr)
			a.Equal(directStored, bufferedStored)
			directBytes, _ := os.ReadFile(direct.Filepath)
			bufferedBytes, _ := os.ReadFile(buffered.Filepath)
			a.Equal(string(directBytes), string(bufferedBytes))
		})
	}
}

// Recorders only read the restored snapshots, so many test cases may compare at
// once; this test exists to be run under -race.
func TestRecorderConcurrentCompare(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	a.Nil(cache.RestoreFromFile())

	recorders := make([]*Recorder, 8)
	var wg sync.WaitGroup
	for i := range recorders {
		recorders[i] = cache.NewRecorder()
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorders[i].Compare(fmt.Sprintf("test %d", i), 1, content1)
		}()
	}
	wg.Wait()
	for _, recorder := range recorders {
		cache.Commit(recorder)
	}

	a.Equal(uint(8), cache.CurrentCount())
	a.Equal(uint(8), cache.InsertedCount())
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/unittest/snapshot/ -run 'TestRecorder' -race`
Expected: build FAIL with `cache.NewRecorder undefined` / `undefined: Recorder`.

- [ ] **Step 3: Create `pkg/unittest/snapshot/recorder.go`**

Move the body of `Cache.Compare` here, rewritten against the recorder's own buffer:

```go
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
```

- [ ] **Step 4: Replace `Cache.Compare` in `cache.go`**

Replace the whole existing `Compare` method body with:

```go
// Compare content to cached last time, return CompareResult
func (s *Cache) Compare(test string, idx uint, content any, optFns ...func(options *CacheOptions) error) *CompareResult {
	recorder := s.NewRecorder()
	result := recorder.Compare(test, idx, content, optFns...)
	s.Commit(recorder)
	return result
}
```

Remove the now-unused `"fmt"` and `".../valueutils"` imports from `cache.go` (keep `bytes`, `os`, `strings`, `internal/common`).

- [ ] **Step 5: Run the snapshot package tests**

Run: `go test ./pkg/unittest/snapshot/ -race`
Expected: PASS (new and all existing `TestCache*` tests).

- [ ] **Step 6: Commit**

```bash
git add pkg/unittest/snapshot/
git commit -m "feat(snapshot): buffer comparisons in a per-case recorder

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Test job writes snapshots through a recorder

**Files:**
- Modify: `pkg/unittest/models.go` (`TestConfig` struct, add option after `WithCoverageTracker`)
- Modify: `pkg/unittest/test_job.go:190-199` (`orderedSnapshotComparer`) and `:316` (its construction)
- Test: `pkg/unittest/test_job_test.go`

**Interfaces:**
- Consumes: `Cache.NewRecorder`, `Recorder.Compare`, `Cache.Commit` (Task 1).
- Produces: `func WithSnapshotRecorder(rec *snapshot.Recorder) LoadTestOptionsFunc` — a nil recorder means "compare straight on the cache".

- [ ] **Step 1: Write the failing test**

Append to `pkg/unittest/test_job_test.go`:

```go
func TestV4RunJobWithSnapshotRecorderDefersCacheWrites(t *testing.T) {
	c, _ := v2loader.Load(testV4BasicChart)
	manifest := `
it: should snapshot
template: templates/configmap.yaml
asserts:
  - matchSnapshot: {}
`
	var tj TestJob
	common.YmlUnmarshalTestHelper(manifest, &tj, t)

	cache := &snapshot.Cache{}
	recorder := cache.NewRecorder()
	tj.WithConfig(*NewTestConfig(c, cache, WithSnapshotRecorder(recorder)))
	testResult := tj.RunV4(&results.TestJobResult{})

	a := assert.New(t)
	a.NoError(testResult.ExecError)
	a.True(testResult.Passed)
	a.Equal(uint(0), cache.CurrentCount(), "snapshots must stay in the recorder until committed")

	cache.Commit(recorder)
	a.Positive(cache.CurrentCount())
	a.Equal(cache.CurrentCount(), cache.InsertedCount())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/unittest/ -run TestV4RunJobWithSnapshotRecorderDefersCacheWrites`
Expected: build FAIL with `undefined: WithSnapshotRecorder`.

- [ ] **Step 3: Add the option in `models.go`**

Add the field as the last field of `TestConfig`:

```go
	snapshotRecorder       *snapshot.Recorder
```

Add after `WithCoverageTracker`:

```go
// WithSnapshotRecorder makes the test job buffer its snapshot comparisons in the
// recorder instead of writing them to the cache; a nil recorder writes to the cache.
func WithSnapshotRecorder(rec *snapshot.Recorder) LoadTestOptionsFunc {
	return func(c *TestConfig) {
		c.snapshotRecorder = rec
	}
}
```

- [ ] **Step 4: Point the comparer at the recorder in `test_job.go`**

Replace the `orderedSnapshotComparer` type and method with:

```go
// snapshotComparison is implemented by snapshot.Cache, which stores a comparison
// right away, and by snapshot.Recorder, which buffers it until it is committed.
type snapshotComparison interface {
	Compare(test string, idx uint, content any, optFns ...func(options *snapshot.CacheOptions) error) *snapshot.CompareResult
}

type orderedSnapshotComparer struct {
	target  snapshotComparison
	test    string
	counter uint
}

func (s *orderedSnapshotComparer) CompareToSnapshot(content any, optFns ...func(options *snapshot.CacheOptions) error) *snapshot.CompareResult {
	s.counter++
	return s.target.Compare(s.test, s.counter, content, optFns...)
}
```

Replace line `snapshotComparer := &orderedSnapshotComparer{cache: t.configOrDefault().cache, test: t.Name}` with:

```go
	var snapshotTarget snapshotComparison = t.configOrDefault().cache
	if recorder := t.configOrDefault().snapshotRecorder; recorder != nil {
		snapshotTarget = recorder
	}
	snapshotComparer := &orderedSnapshotComparer{target: snapshotTarget, test: t.Name}
```

- [ ] **Step 5: Run the package tests**

Run: `go test ./pkg/unittest/ -race`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/unittest/models.go pkg/unittest/test_job.go pkg/unittest/test_job_test.go
git commit -m "feat: let test jobs record snapshots through a recorder

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Fixture chart + failfast must not rewrite snapshots of cases that never ran

Fixes the existing bug from the spec: a failfast stop inside the **last** suite of a group still stores the `.snap` file and deletes the snapshots of the cases that did not run. Applies to all modes.

**Files:**
- Create: `test/data/v3/parallel-testcases/Chart.yaml`
- Create: `test/data/v3/parallel-testcases/values.yaml`
- Create: `test/data/v3/parallel-testcases/templates/configmap.yaml`
- Create: `test/data/v3/parallel-testcases/tests/cases_test.yaml`
- Create: `test/data/v3/parallel-testcases/tests/more_test.yaml`
- Create: `test/data/v3/parallel-testcases/tests/__snapshot__/*.snap` (generated in Step 2)
- Create: `test/data/v3/parallel-testcases/tests_failed/failfast_test.yaml`
- Create: `test/data/v3/parallel-testcases/tests_dup/dup_test.yaml`
- Create: `test/data/v3/parallel-testcases/tests_skipped/skipped_test.yaml`
- Modify: `pkg/unittest/test_runner.go` (`runSuiteGroup` loop ~line 576, new `cutShort` method next to `failFast`, add `"slices"` import)
- Test: `pkg/unittest/test_runner_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces (used by Tasks 4 and 5):
  - Fixture path `../../test/data/v3/parallel-testcases` (from `pkg/unittest`).
  - Test helpers in `test_runner_test.go` (package `unittest_test`):
    - `const testV4ParallelTestCasesChart string = "../../test/data/v3/parallel-testcases"`
    - `func copyFixtureToTemp(t *testing.T, src string) string`
    - `func runFixture(chartDir, testFiles string, configure func(*TestRunner)) (bool, string)` — returns passed and the printed output with times masked.
  - `func (r suiteRunResult) cutShort() bool`

- [ ] **Step 1: Create the fixture**

`Chart.yaml`:

```yaml
apiVersion: v2
name: parallel-testcases
description: A chart with many test cases per suite for parallel test case execution
type: application
version: 0.1.0
appVersion: "1.0.0"
```

`values.yaml`:

```yaml
name: cases
```

`templates/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  name: {{ .Values.name }}
```

`tests/cases_test.yaml` (two suites in one file → one shared snapshot group):

```yaml
suite: configmap cases
templates:
  - templates/configmap.yaml
tests:
  - it: renders default name
    asserts:
      - matchSnapshot: {}
  - it: renders name one
    set:
      name: one
    asserts:
      - equal:
          path: data.name
          value: one
      - matchSnapshot: {}
  - it: renders name two
    set:
      name: two
    asserts:
      - equal:
          path: data.name
          value: two
      - matchSnapshot: {}
  - it: is skipped
    skip:
      reason: demonstrates a skipped case between running ones
    asserts:
      - matchSnapshot: {}
  - it: renders name three
    set:
      name: three
    asserts:
      - matchSnapshot: {}
  - it: renders name four
    set:
      name: four
    asserts:
      - matchSnapshot: {}
---
suite: configmap release cases
templates:
  - templates/configmap.yaml
tests:
  - it: uses release name
    release:
      name: my-release
    asserts:
      - equal:
          path: metadata.name
          value: my-release-config
  - it: snapshots release name
    release:
      name: other-release
    asserts:
      - matchSnapshot: {}
```

`tests/more_test.yaml` (a second snapshot group):

```yaml
suite: more configmap cases
templates:
  - templates/configmap.yaml
tests:
  - it: renders name five
    set:
      name: five
    asserts:
      - matchSnapshot: {}
  - it: renders name six
    set:
      name: six
    asserts:
      - matchSnapshot: {}
  - it: renders name seven
    set:
      name: seven
    asserts:
      - equal:
          path: data.name
          value: seven
```

`tests_failed/failfast_test.yaml`:

```yaml
suite: failfast cases
templates:
  - templates/configmap.yaml
tests:
  - it: passes first
    asserts:
      - matchSnapshot: {}
  - it: fails second
    asserts:
      - equal:
          path: data.name
          value: wrong
  - it: passes third
    set:
      name: third
    asserts:
      - matchSnapshot: {}
  - it: passes fourth
    set:
      name: fourth
    asserts:
      - matchSnapshot: {}
```

`tests_dup/dup_test.yaml`:

```yaml
suite: duplicate names
templates:
  - templates/configmap.yaml
tests:
  - it: same name
    set:
      name: first
    asserts:
      - matchSnapshot: {}
  - it: same name
    set:
      name: second
    asserts:
      - matchSnapshot: {}
  - it: same name
    set:
      name: third
    asserts:
      - matchSnapshot: {}
```

`tests_skipped/skipped_test.yaml`:

```yaml
suite: all skipped
templates:
  - templates/configmap.yaml
tests:
  - it: skipped one
    skip:
      reason: skipped on purpose
    asserts:
      - matchSnapshot: {}
  - it: skipped two
    skip:
      reason: skipped on purpose
    asserts:
      - matchSnapshot: {}
```

- [ ] **Step 2: Generate and commit the passing suites' snapshots**

Run: `go run ./cmd/helm-unittest -u -f 'tests/*_test.yaml' test/data/v3/parallel-testcases`
Expected: all tests pass; `test/data/v3/parallel-testcases/tests/__snapshot__/cases_test.yaml.snap` and `more_test.yaml.snap` are created. Run it again without `-u`; expected: pass, `0 created`.
Do **not** generate snapshots for `tests_failed`, `tests_dup` or `tests_skipped`; tests copy the fixture to a temp dir before writing there.

- [ ] **Step 3: Write the failing test**

Add to `pkg/unittest/test_runner_test.go`:

```go
const testV4ParallelTestCasesChart string = "../../test/data/v3/parallel-testcases"

// copyFixtureToTemp copies a fixture chart into a throwaway directory so runs that
// write snapshot files never touch the repository.
func copyFixtureToTemp(t *testing.T, src string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "chart")
	if err := os.CopyFS(dir, os.DirFS(src)); err != nil {
		t.Fatalf("failed to copy fixture %s: %v", src, err)
	}
	return dir
}

// runFixture runs the test files of one chart with a runner adjusted by configure and
// returns whether it passed plus its printed output with the elapsed time masked.
func runFixture(chartDir, testFiles string, configure func(*TestRunner)) (bool, string) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{testFiles},
	}
	if configure != nil {
		configure(&runner)
	}
	passed := runner.RunV4([]string{chartDir})
	return passed, timePattern.ReplaceAllString(buffer.String(), "${1}XX.XXXms")
}

// A failfast stop inside the last suite of a snapshot group leaves the later cases
// unrun; their snapshots must not be treated as vanished and deleted.
func TestV4RunnerFailfastKeepsSnapshotsOfCasesThatDidNotRun(t *testing.T) {
	a := assert.New(t)
	chartDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)
	snapPath := filepath.Join(chartDir, "tests_failed", "__snapshot__", "failfast_test.yaml.snap")

	runFixture(chartDir, testTestFailedFiles, func(r *TestRunner) { r.UpdateSnapshot = true })
	before, err := os.ReadFile(snapPath)
	a.NoError(err)
	a.Contains(string(before), "passes fourth")

	passed, output := runFixture(chartDir, testTestFailedFiles, func(r *TestRunner) { r.Failfast = true })
	a.False(passed, output)

	after, err := os.ReadFile(snapPath)
	a.NoError(err)
	a.Equal(string(before), string(after), "failfast must not rewrite the snapshot file")
}
```

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./pkg/unittest/ -run TestV4RunnerFailfastKeepsSnapshotsOfCasesThatDidNotRun`
Expected: FAIL on `failfast must not rewrite the snapshot file` (the `after` content lacks `passes third` / `passes fourth`).

- [ ] **Step 5: Implement the fix in `test_runner.go`**

Add `"slices"` to the imports. Add after `failFast()`:

```go
// cutShort reports whether failfast stopped the suite before all of its test cases
// ran; the cases that did not run have no result.
func (r suiteRunResult) cutShort() bool {
	return r.toAppend != nil && slices.Contains(r.toAppend.TestsResult, nil)
}
```

In `runSuiteGroup`, replace the loop condition:

```go
		if suiteResult.failFast() && i < len(group.suites)-1 {
```

with:

```go
		if suiteResult.failFast() && (i < len(group.suites)-1 || suiteResult.cutShort()) {
```

and update the comment above `if completed && ...` to:

```go
	// Storing a cache that only saw part of the group would drop the snapshots of the
	// suites or test cases that never ran, so a group cut short by --failfast is left
	// untouched.
```

- [ ] **Step 6: Run the package tests**

Run: `go test ./pkg/unittest/... -race`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add test/data/v3/parallel-testcases pkg/unittest/test_runner.go pkg/unittest/test_runner_test.go
git commit -m "fix: keep snapshots of test cases skipped by failfast

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Run the test cases of a suite on a case pool

**Files:**
- Modify: `pkg/unittest/test_suite.go` (`TestSuite` fields ~line 245-256; `runV4TestJobs` lines 395-446; imports `sync`, `sync/atomic`)
- Create: `pkg/unittest/test_suite_parallel_internal_test.go`

**Interfaces:**
- Consumes: `Cache.NewRecorder`, `Cache.Commit` (Task 1); `WithSnapshotRecorder` (Task 2); fixture (Task 3).
- Produces (used by Task 5):
  - `TestSuite.casePool chan struct{}` — nil = sequential; set = run cases on this pool.
  - `TestSuite.caseStartHook func()` — test-only hook, called after a case acquired its slot.

- [ ] **Step 1: Write the failing tests**

Create `pkg/unittest/test_suite_parallel_internal_test.go`:

```go
package unittest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	v2chart "helm.sh/helm/v4/pkg/chart/v2"
	v2loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

const internalTestCasesChart = "../../test/data/v3/parallel-testcases"

// runFixtureSuite parses the first suite of a fixture test file and runs it against
// a fresh snapshot cache, on a case pool of the given size when workers > 0.
func runFixtureSuite(t *testing.T, suiteFile string, workers int, failFast, update bool, hook func()) (*results.TestSuiteResult, *snapshot.Cache) {
	t.Helper()
	suite, chart, cache := prepareFixtureSuite(t, suiteFile, workers, update, hook)
	return suite.RunV4(chart, cache, failFast, "", &results.TestSuiteResult{}), cache
}

// prepareFixtureSuite loads the fixture chart and the first suite of a test file. It
// must be called from the test goroutine because it may stop the test.
func prepareFixtureSuite(t *testing.T, suiteFile string, workers int, update bool, hook func()) (*TestSuite, *v2chart.Chart, *snapshot.Cache) {
	t.Helper()
	chart, err := v2loader.Load(internalTestCasesChart)
	if err != nil {
		t.Fatalf("failed to load fixture chart: %v", err)
	}
	suites, err := ParseTestSuiteFile(filepath.Join(internalTestCasesChart, suiteFile), chart.Name(), true, nil)
	if err != nil || len(suites) == 0 {
		t.Fatalf("failed to parse %s: %v", suiteFile, err)
	}
	suite := suites[0]
	if workers > 0 {
		suite.casePool = make(chan struct{}, workers)
	}
	suite.caseStartHook = hook
	cache := &snapshot.Cache{Filepath: filepath.Join(t.TempDir(), "suite.snap"), IsUpdating: update}
	return suite, chart, cache
}

// describeSuiteResult lists everything a suite run reports, leaving out durations.
func describeSuiteResult(r *results.TestSuiteResult) []string {
	lines := []string{fmt.Sprintf("passed=%v failfast=%v skipped=%v snapshots=%+v",
		r.Passed, r.FailFast, r.Skipped, r.SnapshotCounting)}
	for idx, job := range r.TestsResult {
		if job == nil {
			lines = append(lines, fmt.Sprintf("%d: not run", idx))
			continue
		}
		lines = append(lines, fmt.Sprintf("%d: %s passed=%v skipped=%v err=%v asserts=%d",
			idx, job.DisplayName, job.Passed, job.Skipped, job.ExecError, len(job.AssertsResult)))
	}
	return lines
}

// storedSnapshot stores the cache and returns the written file content.
func storedSnapshot(t *testing.T, cache *snapshot.Cache) string {
	t.Helper()
	if _, err := cache.StoreToFileIfNeeded(); err != nil {
		t.Fatalf("failed to store snapshot: %v", err)
	}
	content, err := os.ReadFile(cache.Filepath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	return string(content)
}

func TestRunV4TestJobsParallelMatchesSequential(t *testing.T) {
	a := assert.New(t)
	seq, seqCache := runFixtureSuite(t, "tests/cases_test.yaml", 0, false, true, nil)
	par, parCache := runFixtureSuite(t, "tests/cases_test.yaml", 3, false, true, nil)

	a.Equal(describeSuiteResult(seq), describeSuiteResult(par))
	seqSnap := storedSnapshot(t, seqCache)
	a.NotEmpty(seqSnap)
	a.Equal(seqSnap, storedSnapshot(t, parCache))
}

func TestRunV4TestJobsParallelFailfastMatchesSequential(t *testing.T) {
	a := assert.New(t)
	seq, seqCache := runFixtureSuite(t, "tests_failed/failfast_test.yaml", 0, true, false, nil)
	par, parCache := runFixtureSuite(t, "tests_failed/failfast_test.yaml", 4, true, false, nil)

	a.Equal(describeSuiteResult(seq), describeSuiteResult(par))
	a.Nil(par.TestsResult[2], "cases after the failure must not be reported")
	a.Equal(seqCache.CurrentCount(), parCache.CurrentCount())
	a.Equal(seqCache.InsertedCount(), parCache.InsertedCount())
	a.Equal(seqCache.VanishedCount(), parCache.VanishedCount())
}

func TestRunV4TestJobsParallelDuplicateNamesMatchSequential(t *testing.T) {
	a := assert.New(t)
	_, seqCache := runFixtureSuite(t, "tests_dup/dup_test.yaml", 0, false, true, nil)
	seqSnap := storedSnapshot(t, seqCache)
	a.Contains(seqSnap, "third", "the last duplicate wins")

	for run := 0; run < 5; run++ {
		_, parCache := runFixtureSuite(t, "tests_dup/dup_test.yaml", 3, false, true, nil)
		a.Equal(seqSnap, storedSnapshot(t, parCache), "run %d", run)
	}
}

func TestRunV4TestJobsParallelAllSkipped(t *testing.T) {
	a := assert.New(t)
	seq, seqCache := runFixtureSuite(t, "tests_skipped/skipped_test.yaml", 0, false, false, nil)
	par, parCache := runFixtureSuite(t, "tests_skipped/skipped_test.yaml", 2, false, false, nil)

	a.Equal(describeSuiteResult(seq), describeSuiteResult(par))
	a.True(par.Skipped)
	a.True(par.Passed)
	a.Equal(uint(0), parCache.CurrentCount())
	a.Equal(seqCache.CurrentCount(), parCache.CurrentCount())
}

// With a pool of two slots, two cases start and block in the hook; a third must not
// start until one of them is released.
func TestRunV4TestJobsParallelRespectsPoolSize(t *testing.T) {
	const workers = 2
	started := make(chan struct{}, 16)
	release := make(chan struct{})
	hook := func() {
		started <- struct{}{}
		<-release
	}

	suite, chart, cache := prepareFixtureSuite(t, "tests/cases_test.yaml", workers, true, hook)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		suite.RunV4(chart, cache, false, "", &results.TestSuiteResult{})
	}()

	for i := 0; i < workers; i++ {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			<-finished
			t.Fatalf("only %d cases ran concurrently, expected %d", i, workers)
		}
	}
	select {
	case <-started:
		close(release)
		<-finished
		t.Fatalf("more than %d cases ran concurrently", workers)
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	<-finished
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/unittest/ -run 'TestRunV4TestJobsParallel' -race`
Expected: build FAIL with `suite.casePool undefined`.

- [ ] **Step 3: Add the fields to `TestSuite`**

After the `coverageTracker` field:

```go
	// casePool, when set, runs the test cases concurrently: each case holds one slot
	// of this pool, shared by all suites of the run, while it executes.
	casePool chan struct{} `yaml:"-"`
	// caseStartHook is called when a test case starts on the case pool. It is nil in
	// normal operation and exists only so tests can observe case concurrency.
	caseStartHook func() `yaml:"-"`
```

- [ ] **Step 4: Rewrite `runV4TestJobs` in `test_suite.go`**

Replace the whole existing `runV4TestJobs` function with the following (keep `SuiteResult` as is; add `"sync"` and `"sync/atomic"` imports):

```go
func (s *TestSuite) runV4TestJobs(
	chart *v2chart.Chart,
	cache *snapshot.Cache,
	failFast bool,
	renderPath string,
	tracker *coverage.Tracker,
) *SuiteResult {
	if s.casePool != nil {
		return s.runV4TestJobsParallel(chart, cache, failFast, renderPath, tracker)
	}

	acc := newJobAccumulator(len(s.Tests))
	for idx, testJob := range s.Tests {
		jobResult := s.runV4TestJob(idx, testJob, chart, cache, nil, failFast, renderPath, tracker)
		if acc.add(idx, jobResult, testJob.Skip.Reason != "", failFast) {
			break
		}
	}
	return acc.finish()
}

// runV4TestJobsParallel runs the test cases on the case pool and then reports them in
// index order, so results and snapshots equal a sequential run. Under failfast, cases
// after the first failing one are not started once the failure is known, and are
// dropped from the report and the snapshot cache if they already ran.
func (s *TestSuite) runV4TestJobsParallel(
	chart *v2chart.Chart,
	cache *snapshot.Cache,
	failFast bool,
	renderPath string,
	tracker *coverage.Tracker,
) *SuiteResult {
	type caseOutcome struct {
		result   *results.TestJobResult
		recorder *snapshot.Recorder
	}
	outcomes := make([]caseOutcome, len(s.Tests))

	// Index of the lowest failing case under failfast; cases above it do not start.
	var stopAfter atomic.Int64
	stopAfter.Store(int64(len(s.Tests)))

	var wg sync.WaitGroup
	for idx, testJob := range s.Tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.casePool <- struct{}{}
			defer func() { <-s.casePool }()

			if int64(idx) > stopAfter.Load() {
				return
			}
			if s.caseStartHook != nil {
				s.caseStartHook()
			}

			recorder := cache.NewRecorder()
			jobResult := s.runV4TestJob(idx, testJob, chart, cache, recorder, failFast, renderPath, tracker)
			outcomes[idx] = caseOutcome{result: jobResult, recorder: recorder}
			if failFast && testJob.Skip.Reason == "" && !jobResult.Passed {
				lowerAtomic(&stopAfter, int64(idx))
			}
		}()
	}
	wg.Wait()

	acc := newJobAccumulator(len(s.Tests))
	for idx, outcome := range outcomes {
		if outcome.result == nil {
			break
		}
		cache.Commit(outcome.recorder)
		if acc.add(idx, outcome.result, s.Tests[idx].Skip.Reason != "", failFast) {
			break
		}
	}
	return acc.finish()
}

// lowerAtomic stores n in v when n is lower than the current value.
func lowerAtomic(v *atomic.Int64, n int64) {
	for {
		current := v.Load()
		if n >= current || v.CompareAndSwap(current, n) {
			return
		}
	}
}

// runV4TestJob runs the test case at idx against its own deep copy of the chart. A
// case skipped by reason is not rendered. Snapshot comparisons go to the recorder
// when one is given, otherwise straight to the cache.
func (s *TestSuite) runV4TestJob(
	idx int,
	testJob *TestJob,
	chart *v2chart.Chart,
	cache *snapshot.Cache,
	recorder *snapshot.Recorder,
	failFast bool,
	renderPath string,
	tracker *coverage.Tracker,
) *results.TestJobResult {
	job := results.TestJobResult{DisplayName: testJob.Name, Index: idx}
	if testJob.Skip.Reason != "" {
		job.Skipped = true
		return &job
	}

	chartClone := FullCopyV2Chart(s.chartRoute, chart.Name(), chart)
	testJob.WithConfig(*NewTestConfig(chartClone, cache,
		WithRenderPath(renderPath),
		WithFailFast(failFast),
		WithPostRendererConfig(s.PostRendererConfig),
		WithDocumentSelector(testJob.DocumentSelector),
		WithIncludeCrds(s.IncludeCrds),
		WithSkipSchemaValidation(s.skipSchemaValidation),
		WithCoverageTracker(tracker),
		WithSnapshotRecorder(recorder),
	))
	return testJob.RunV4(&job)
}

// jobAccumulator folds test case results into a SuiteResult in index order. It holds
// the rules of the sequential loop so the parallel path reports exactly the same.
type jobAccumulator struct {
	result  SuiteResult
	skipped int
}

func newJobAccumulator(total int) *jobAccumulator {
	return &jobAccumulator{result: SuiteResult{JobResults: make([]*results.TestJobResult, total)}}
}

// add records the result of the case at idx and reports whether failfast stops the
// suite at this case.
func (a *jobAccumulator) add(idx int, jobResult *results.TestJobResult, skippedByReason, failFast bool) bool {
	a.result.JobResults[idx] = jobResult
	if skippedByReason {
		a.skipped++
		if idx == 0 {
			a.result.Pass = true
		}
	} else {
		if idx == 0 {
			a.result.Pass = jobResult.Passed
		}
		a.result.Pass = a.result.Pass && jobResult.Passed
	}

	if !a.result.Pass && failFast {
		a.result.FailFast = true
		return true
	}
	return false
}

func (a *jobAccumulator) finish() *SuiteResult {
	a.result.Skip = a.skipped == len(a.result.JobResults)
	return &a.result
}
```

- [ ] **Step 5: Run the package tests**

Run: `go test ./pkg/unittest/... -race`
Expected: PASS — the new tests and every existing suite/runner test (the sequential path now goes through `jobAccumulator` and must behave identically).

- [ ] **Step 6: Commit**

```bash
git add pkg/unittest/test_suite.go pkg/unittest/test_suite_parallel_internal_test.go
git commit -m "feat: run the test cases of a suite on a shared case pool

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Runner dispatch for `ParallelLevel`

**Files:**
- Modify: `pkg/unittest/test_runner.go` (constants, `TestRunner` fields, `runV4SuitesOfChart` ~505-512, `runV4SuitesOfChartSequential` ~533, `runSuiteGroup`, `runSingleSuite`, `runV4SuitesOfChartParallel` ~677-736)
- Test: `pkg/unittest/test_runner_test.go`, `pkg/unittest/test_runner_internal_test.go`

**Interfaces:**
- Consumes: `TestSuite.casePool`, `TestSuite.caseStartHook` (Task 4); `copyFixtureToTemp`, `runFixture`, `testV4ParallelTestCasesChart` (Task 3).
- Produces (used by Task 6):
  - `const ParallelLevelSuite = "suite"`, `const ParallelLevelTest = "test"`
  - `TestRunner.ParallelLevel string` — empty means `suite`.
  - `TestRunner.caseStartHook func()` (unexported, tests only).

- [ ] **Step 1: Write the failing runner tests**

Append to `pkg/unittest/test_runner_test.go`:

```go
func testLevel(r *TestRunner) {
	r.Parallel = true
	r.ParallelLevel = ParallelLevelTest
	r.MaxWorkers = 3
}

func TestV4RunnerTestLevelDeterministicOutput(t *testing.T) {
	chartDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)

	_, sequential := runFixture(chartDir, testTestFiles, nil)
	_, first := runFixture(chartDir, testTestFiles, testLevel)
	_, second := runFixture(chartDir, testTestFiles, testLevel)

	assert.Equal(t, first, second, "test-level output must be stable across runs")
	assert.Equal(t, sequential, first, "test-level output must match sequential")

	_, basicSequential := runFixture(testV4BasicChart, testTestFiles, nil)
	_, basicTestLevel := runFixture(testV4BasicChart, testTestFiles, testLevel)
	assert.Equal(t, basicSequential, basicTestLevel)
}

func TestV4RunnerTestLevelUpdateSnapshotMatchesSequential(t *testing.T) {
	for _, testFiles := range []string{testTestFiles, "tests_dup/*_test.yaml"} {
		t.Run(testFiles, func(t *testing.T) {
			seqDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)
			parDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)
			// Drop committed snapshots so both runs write every file from scratch.
			assert.NoError(t, os.RemoveAll(filepath.Join(seqDir, "tests", "__snapshot__")))
			assert.NoError(t, os.RemoveAll(filepath.Join(parDir, "tests", "__snapshot__")))

			seqPassed, seqOut := runFixture(seqDir, testFiles, func(r *TestRunner) { r.UpdateSnapshot = true })
			parPassed, parOut := runFixture(parDir, testFiles, func(r *TestRunner) {
				testLevel(r)
				r.UpdateSnapshot = true
			})
			assert.True(t, seqPassed, seqOut)
			assert.Equal(t, seqPassed, parPassed)
			assert.Equal(t, seqOut, parOut)

			// "tests/*_test.yaml" -> "tests", "tests_dup/*_test.yaml" -> "tests_dup".
			snapDir := filepath.Dir(testFiles)
			seqSnaps, err := filepath.Glob(filepath.Join(seqDir, snapDir, "__snapshot__", "*.snap"))
			assert.NoError(t, err)
			assert.NotEmpty(t, seqSnaps)
			for _, seqSnap := range seqSnaps {
				rel, _ := filepath.Rel(seqDir, seqSnap)
				seqBytes, _ := os.ReadFile(seqSnap)
				parBytes, err := os.ReadFile(filepath.Join(parDir, rel))
				assert.NoError(t, err)
				assert.Equal(t, string(seqBytes), string(parBytes), rel)
			}
		})
	}
}

func TestV4RunnerTestLevelFailfastMatchesSequential(t *testing.T) {
	seqDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)
	parDir := copyFixtureToTemp(t, testV4ParallelTestCasesChart)

	seqPassed, seqOut := runFixture(seqDir, testTestFailedFiles, func(r *TestRunner) { r.Failfast = true })
	parPassed, parOut := runFixture(parDir, testTestFailedFiles, func(r *TestRunner) {
		testLevel(r)
		r.Failfast = true
	})

	assert.False(t, seqPassed)
	assert.Equal(t, seqPassed, parPassed)
	assert.Equal(t, seqOut, parOut)

	snap := filepath.Join("tests_failed", "__snapshot__", "failfast_test.yaml.snap")
	_, seqErr := os.Stat(filepath.Join(seqDir, snap))
	_, parErr := os.Stat(filepath.Join(parDir, snap))
	assert.Equal(t, os.IsNotExist(seqErr), os.IsNotExist(parErr), "snapshot file presence must match")
}

func TestV4RunnerTestLevelCoverageMatchesSequential(t *testing.T) {
	_, sequential := runFixture(testV4BasicChart, testTestFiles, func(r *TestRunner) { r.Coverage = true })
	_, testLevelOut := runFixture(testV4BasicChart, testTestFiles, func(r *TestRunner) {
		testLevel(r)
		r.Coverage = true
	})
	assert.Equal(t, sequential, testLevelOut)
}
```

Append to `pkg/unittest/test_runner_internal_test.go` (add `"sync/atomic"` import):

```go
const internalTestCasesRunnerChart = "../../test/data/v3/parallel-testcases"

// At test level the cases of every suite share one pool: the peak number of cases
// running at once across the whole chart equals MaxWorkers.
func TestRunV4TestLevelSharesOnePoolAcrossSuites(t *testing.T) {
	const workers = 2
	var running, peak atomic.Int32

	buffer := new(bytes.Buffer)
	tr := &TestRunner{
		Printer:       printer.NewPrinter(buffer, nil),
		TestFiles:     []string{"tests/*_test.yaml"},
		Parallel:      true,
		ParallelLevel: ParallelLevelTest,
		MaxWorkers:    workers,
	}
	tr.caseStartHook = func() {
		now := running.Add(1)
		for {
			highest := peak.Load()
			if now <= highest || peak.CompareAndSwap(highest, now) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		running.Add(-1)
	}

	assert.True(t, tr.RunV4([]string{internalTestCasesRunnerChart}), buffer.String())
	assert.Equal(t, int32(workers), peak.Load())
}

// --debugPlugin writes rendered files keyed by template path, so test level must fall
// back to sequential execution and never use the case pool.
func TestRunV4TestLevelIgnoredWithDebugRender(t *testing.T) {
	var calls atomic.Int32
	buffer := new(bytes.Buffer)
	tr := &TestRunner{
		Printer:       printer.NewPrinter(buffer, nil),
		TestFiles:     []string{"tests/*_test.yaml"},
		Parallel:      true,
		ParallelLevel: ParallelLevelTest,
		RenderPath:    t.TempDir(),
	}
	tr.caseStartHook = func() { calls.Add(1) }

	tr.RunV4([]string{internalTestCasesRunnerChart})
	assert.Equal(t, int32(0), calls.Load())
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./pkg/unittest/ -run 'TestLevel' -race`
Expected: build FAIL with `undefined: ParallelLevelTest` / `unknown field ParallelLevel`.

- [ ] **Step 3: Add constants and fields to `test_runner.go`**

After `const LOG_TEST_RUNNER = "test-runner"`:

```go
// Granularity of parallel execution, see TestRunner.ParallelLevel.
const (
	// ParallelLevelSuite runs test suites concurrently.
	ParallelLevelSuite = "suite"
	// ParallelLevelTest runs the test cases of all suites concurrently in one pool.
	ParallelLevelTest = "test"
)
```

In `TestRunner`, after `MaxWorkers int`:

```go
	ParallelLevel        string
```

After the `suiteStartHook` field:

```go
	// caseStartHook is passed to every suite at test level; see TestSuite.caseStartHook.
	caseStartHook func()
```

- [ ] **Step 4: Dispatch and thread the pool through**

Replace `runV4SuitesOfChart`:

```go
// runV4SuitesOfChart runs suite files of the chart and print output.
// It dispatches to the parallel path when enabled (and not in debug render mode),
// otherwise runs sequentially with identical output to the historical behavior.
func (tr *TestRunner) runV4SuitesOfChart(suites []*TestSuite, chart *v2chart.Chart) bool {
	if tr.Parallel && tr.RenderPath == "" {
		groups := groupSuitesBySnapshotFile(suites)
		if tr.ParallelLevel == ParallelLevelTest {
			// Groups only coordinate; the test cases of all suites share one pool.
			pool := make(chan struct{}, tr.effectiveWorkers(countTestCases(suites)))
			return tr.runV4SuiteGroupsParallel(groups, chart, len(groups), pool)
		}
		return tr.runV4SuiteGroupsParallel(groups, chart, tr.effectiveWorkers(len(groups)), nil)
	}
	return tr.runV4SuitesOfChartSequential(suites, chart)
}

// countTestCases returns the number of test cases of all suites.
func countTestCases(suites []*TestSuite) int {
	total := 0
	for _, suite := range suites {
		total += len(suite.Tests)
	}
	return total
}
```

In `runV4SuitesOfChartSequential` change `tr.runSuiteGroup(group, chart)` to `tr.runSuiteGroup(group, chart, nil)`.

Change `runSuiteGroup`'s signature and its `runSingleSuite` call:

```go
func (tr *TestRunner) runSuiteGroup(group *suiteGroup, chart *v2chart.Chart, casePool chan struct{}) []suiteRunResult {
```

```go
		suiteResult := tr.runSingleSuite(suite, chart, snapshotCache, casePool)
```

Add to its doc comment: `When casePool is set, the test cases of each suite run on that pool.`

Change `runSingleSuite`:

```go
// runSingleSuite runs one suite against the snapshot cache of its snapshot file and
// returns its results. A non-nil casePool runs the suite's test cases on that pool.
func (tr *TestRunner) runSingleSuite(suite *TestSuite, chart *v2chart.Chart, snapshotCache *snapshot.Cache, casePool chan struct{}) suiteRunResult {
	if tr.suiteStartHook != nil {
		tr.suiteStartHook()
	}

	snapshotCache.BeginSuite()
	suite.skipSchemaValidation = tr.SkipSchemaValidation
	suite.casePool = casePool
	suite.caseStartHook = tr.caseStartHook
	result := suite.RunV4(chart, snapshotCache, tr.Failfast, tr.RenderPath, &results.TestSuiteResult{})

	return suiteRunResult{
		toPrint:  []*results.TestSuiteResult{result},
		toAppend: result,
	}
}
```

Rename `runV4SuitesOfChartParallel` to `runV4SuiteGroupsParallel` and change it to:

```go
// runV4SuiteGroupsParallel runs snapshot-file groups concurrently, at most groupLimit
// at a time, while keeping all printing, counting and result collection on the main
// goroutine in discovery order. A non-nil casePool also runs the test cases of every
// suite on that pool; every group then starts at once, so printing stops where the
// sequential loop stops under --failfast.
func (tr *TestRunner) runV4SuiteGroupsParallel(groups []*suiteGroup, chart *v2chart.Chart, groupLimit int, casePool chan struct{}) bool {
	if len(groups) == 0 {
		return true
	}

	// Results land in disjoint, preallocated slots so no locking is needed.
	groupResults := make([][]suiteRunResult, len(groups))
	done := make([]chan struct{}, len(groups))
	for i := range done {
		done[i] = make(chan struct{})
	}

	var failed atomic.Bool
	semaphore := make(chan struct{}, groupLimit)
	var wg sync.WaitGroup

	for i, group := range groups {
		wg.Add(1)
		go func(idx int, g *suiteGroup) {
			defer wg.Done()
			defer close(done[idx])
			defer func() { <-semaphore }()
			semaphore <- struct{}{}

			// Failfast: once a suite has failed, skip scheduling further groups.
			// Leaving results nil mirrors the sequential break which leaves later
			// suites unreported.
			if tr.Failfast && failed.Load() {
				return
			}

			out := tr.runSuiteGroup(g, chart, casePool)
			for _, suiteResult := range out {
				// Trigger failfast only on the primary result, matching the
				// sequential break; synthesized snapshot errors must not stop
				// scheduling of later groups.
				if suiteResult.failFast() {
					failed.Store(true)
				}
			}
			groupResults[idx] = out
		}(i, group)
	}

	// Stream output deterministically: wait on each group in order, then print.
	stopOnFailfast := casePool != nil && tr.Failfast
	chartPassed := true
printing:
	for i := range groups {
		<-done[i]
		for _, suiteResult := range groupResults[i] {
			for _, result := range suiteResult.toPrint {
				chartPassed = chartPassed && result.Passed
				tr.handleSuiteResult(result)
			}
			if suiteResult.toAppend != nil {
				tr.testResults = append(tr.testResults, suiteResult.toAppend)
			}

			if stopOnFailfast && !chartPassed && suiteResult.failFast() {
				break printing
			}
		}
	}

	wg.Wait()
	return chartPassed
}
```

- [ ] **Step 5: Run the package tests**

Run: `go test ./pkg/unittest/... -race`
Expected: PASS — new tests plus all existing `--parallel` tests (`TestRunV4SuitesParallelEngagesConcurrency`, `TestParallelFailfastIgnoresSnapshotErrors`, `TestV4RunnerParallel*`) unchanged.

- [ ] **Step 6: Commit**

```bash
git add pkg/unittest/test_runner.go pkg/unittest/test_runner_test.go pkg/unittest/test_runner_internal_test.go
git commit -m "feat: run test cases of all suites in one pool at parallel level test

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: `--parallel-level` CLI flag and docs

**Files:**
- Modify: `cmd/helm-unittest/helm_unittest.go` (`testOptions` struct ~line 29-31; validation near line 101; runner literal ~line 132; flag registration after `--max-workers` ~line 239)
- Test: `cmd/helm-unittest/helm_unittest_test.go`
- Modify: `README.md` (flag list ~line 217-218, parallel section ~line 221-245)
- Modify: `CHANGELOG.md` (`1.3.0` section)

**Interfaces:**
- Consumes: `unittest.ParallelLevelSuite`, `unittest.ParallelLevelTest`, `TestRunner.ParallelLevel` (Task 5).
- Produces: CLI flag `--parallel-level`.

- [ ] **Step 1: Write the failing tests**

Append to `cmd/helm-unittest/helm_unittest_test.go` (add imports `"slices"`, `"strings"`, `log "github.com/sirupsen/logrus"`, `logtest "github.com/sirupsen/logrus/hooks/test"`):

```go
func TestValidateUnittestParallelLevelFlag(t *testing.T) {
	a := assert.New(t)

	levels := map[string]string{
		"":                       "suite",
		"--parallel-level=suite": "suite",
		"--parallel-level=test":  "test",
	}
	for levelFlag, expected := range levels {
		cmd := setupTestCmd()
		args := []string{"--parallel"}
		if levelFlag != "" {
			args = append(args, levelFlag)
		}
		cmd.SetArgs(args)

		err := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(err, "flag: %q", levelFlag)
		a.Equal(expected, runner.ParallelLevel, "flag: %q", levelFlag)
	}
}

func TestValidateUnittestParallelLevelFlagRejectsUnknownValue(t *testing.T) {
	a := assert.New(t)

	cmd := setupTestCmd()
	cmd.SetArgs([]string{"--parallel", "--parallel-level", "bogus"})
	err := cmd.Execute()

	if a.Error(err) {
		a.Contains(err.Error(), "parallel-level")
	}
}

func TestValidateUnittestParallelLevelWithoutParallelWarns(t *testing.T) {
	a := assert.New(t)
	hook := logtest.NewGlobal()
	defer hook.Reset()

	cmd := setupTestCmd()
	cmd.SetArgs([]string{"--parallel-level=test"})
	a.Nil(cmd.Execute())

	a.True(slices.ContainsFunc(hook.AllEntries(), func(entry *log.Entry) bool {
		return entry.Level == log.WarnLevel && strings.Contains(entry.Message, "--parallel-level")
	}), "expected a warning about --parallel-level without --parallel")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/helm-unittest/ -run ParallelLevel`
Expected: FAIL — `unknown flag: --parallel-level`.

- [ ] **Step 3: Implement the flag**

In `testOptions`, after `maxWorkers int`:

```go
	parallelLevel           string
```

After the existing `--debugPlugin` / `--parallel` warning block (around line 101-103), add:

```go
	if testConfig.parallelLevel != unittest.ParallelLevelSuite && testConfig.parallelLevel != unittest.ParallelLevelTest {
		return fmt.Errorf("invalid --parallel-level %q: must be %q or %q",
			testConfig.parallelLevel, unittest.ParallelLevelSuite, unittest.ParallelLevelTest)
	}

	if testConfig.parallelLevel == unittest.ParallelLevelTest && !testConfig.useParallel {
		log.Warn("--parallel-level=test has no effect without --parallel; running sequentially")
	}
```

In the `unittest.TestRunner{...}` literal, after `MaxWorkers:`:

```go
		ParallelLevel:        testConfig.parallelLevel,
```

After the `--max-workers` flag registration:

```go
	cmd.PersistentFlags().StringVar(
		&testConfig.parallelLevel, "parallel-level", unittest.ParallelLevelSuite,
		"granularity of --parallel: 'suite' runs test suites concurrently, 'test' runs the test cases of all suites in one shared worker pool (only used with --parallel)",
	)
```

- [ ] **Step 4: Run the CLI tests**

Run: `go test ./cmd/helm-unittest/ -race`
Expected: PASS.

- [ ] **Step 5: Update README.md**

In the flag list, after the `--max-workers` line:

```
      --parallel-level string   granularity of --parallel: 'suite' runs test suites concurrently, 'test' runs the test cases of all suites in one shared worker pool (only used with --parallel) (default suite)
```

After the paragraph ending "`--max-workers` has no effect unless `--parallel` is also set, and a value of `0` (the default) means one worker per CPU core.", add:

````markdown
By default `--parallel` runs whole test suites concurrently. When a chart has a few
large suites, or many suites sharing one snapshot file, use `--parallel-level test`
to run individual test cases concurrently instead:

```
$ helm unittest --parallel --parallel-level test ./charts/my-app
```

At the `test` level the test cases of all suites share one pool of `--max-workers`
workers, so no more than that many test cases render at once. Output, counts and
snapshot files are identical to a sequential run: results are reported in the order
the tests are defined, and snapshots are written once per snapshot file after all of
its suites finished. With `--failfast`, test cases after the first failing one are
not reported, counted or written to snapshot files, even if they already ran.
`--parallel-level` has no effect without `--parallel`, and `--debugPlugin` still
forces sequential execution.
````

- [ ] **Step 6: Update CHANGELOG.md**

In the `1.3.0 / 2026-10-xx` section, after the `Add code coverage` line:

```
- Add opt-in parallel execution of test cases with `--parallel-level test` (credits @michal-marszalek-h2oai)
- Fix `--failfast` deleting snapshots of test cases that did not run (credits @michal-marszalek-h2oai)
```

- [ ] **Step 7: Full verification**

Run: `make unittest`
Expected: PASS with `-race`, no data race reports.

Run: `go run ./cmd/helm-unittest --parallel --parallel-level test -f 'tests/*_test.yaml' test/data/v3/parallel-testcases test/data/v3/basic`
Expected: all suites pass; output order matches `go run ./cmd/helm-unittest -f 'tests/*_test.yaml' test/data/v3/parallel-testcases test/data/v3/basic`.

Run: `go run ./cmd/helm-unittest --parallel --parallel-level bogus test/data/v3/basic`
Expected: exits non-zero with `invalid --parallel-level "bogus"`.

- [ ] **Step 8: Commit**

```bash
git add cmd/helm-unittest README.md CHANGELOG.md
git commit -m "feat: add --parallel-level flag for parallel test case execution

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
