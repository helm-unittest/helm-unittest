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
func runFixtureSuite(t *testing.T, suiteFile string, workers int, failFast, update bool, hook func(string)) (*results.TestSuiteResult, *snapshot.Cache) {
	t.Helper()
	suite, chart, cache := prepareFixtureSuite(t, suiteFile, workers, update, hook)
	return suite.RunV4(chart, cache, failFast, "", &results.TestSuiteResult{}), cache
}

// prepareFixtureSuite loads the fixture chart and the first suite of a test file. It
// must be called from the test goroutine because it may stop the test.
func prepareFixtureSuite(t *testing.T, suiteFile string, workers int, update bool, hook func(string)) (*TestSuite, *v2chart.Chart, *snapshot.Cache) {
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
	hook := func(string) {
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
