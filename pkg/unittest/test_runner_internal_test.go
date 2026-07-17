package unittest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

const internalTestBasicChart = "../../test/data/v3/basic"
const internalTestMultiSuiteChart = "../../test/data/v3/parallel-multisuite"

// copyChartToTemp copies a fixture chart into a throwaway directory.
func copyChartToTemp(t *testing.T, src string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "chart")
	if err := os.CopyFS(dir, os.DirFS(src)); err != nil {
		t.Fatalf("failed to copy fixture %s: %v", src, err)
	}
	return dir
}

// blockSnapshotDir forces snapshot cache creation to fail by replacing the
// __snapshot__ path with a regular file instead of a directory.
func blockSnapshotDir(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("failed to remove snapshot dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not a dir"), 0644); err != nil {
		t.Fatalf("failed to write blocker file: %v", err)
	}
}

// Snapshot cache/store errors must be printed and counted but kept out of
// tr.testResults, so they never leak into the JUnit/XML formatter output. This
// mirrors the historical sequential behavior on both the sequential and parallel
// paths.
func TestRunSingleSuiteSnapshotErrorNotAppendedToResults(t *testing.T) {
	for _, parallel := range []bool{false, true} {
		name := "sequential"
		if parallel {
			name = "parallel"
		}
		t.Run(name, func(t *testing.T) {
			chartDir := copyChartToTemp(t, internalTestMultiSuiteChart)
			blockSnapshotDir(t, filepath.Join(chartDir, "tests", "__snapshot__"))

			buffer := new(bytes.Buffer)
			tr := &TestRunner{
				Printer:   printer.NewPrinter(buffer, nil),
				TestFiles: []string{"tests/*_test.yaml"},
				Parallel:  parallel,
			}
			passed := tr.RunV3([]string{chartDir})

			assert.False(t, passed, buffer.String())
			// The suites errored: they are counted...
			assert.Positive(t, tr.suiteCounting.failed, "errored suites must still be counted")
			assert.Contains(t, buffer.String(), "is not a directory")
			// ...but the error-synthesized results are not exposed to the formatter.
			assert.Empty(t, tr.testResults, "snapshot cache errors must not be appended to testResults")
		})
	}
}

// The parallel dispatch must actually run suites concurrently. Each suite blocks in
// suiteStartHook until released; if the parallel path engages, MaxWorkers suites reach
// the hook before any is released. A sequential path would start only one and time out.
func TestRunV3SuitesParallelEngagesConcurrency(t *testing.T) {
	const workers = 3

	started := make(chan struct{}, 64)
	release := make(chan struct{})
	suiteStartHook = func() {
		started <- struct{}{}
		<-release
	}
	defer func() { suiteStartHook = nil }()

	buffer := new(bytes.Buffer)
	tr := &TestRunner{
		Printer:    printer.NewPrinter(buffer, nil),
		TestFiles:  []string{"tests/*_test.yaml"},
		Parallel:   true,
		MaxWorkers: workers,
	}

	finished := make(chan bool, 1)
	go func() { finished <- tr.RunV3([]string{internalTestBasicChart}) }()

	for i := 0; i < workers; i++ {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(release)
			<-finished
			t.Fatalf("only %d suites ran concurrently, expected %d", i, workers)
		}
	}
	close(release)
	<-finished
}

// Under --parallel --failfast, a snapshot cache-creation error (a synthesized,
// non-primary result) must not stop later groups from being scheduled, matching
// the sequential path which only breaks on the primary result's FailFast.
func TestParallelFailfastIgnoresSnapshotErrors(t *testing.T) {
	// Every suite in the chart shares one tests/__snapshot__ dir; blocking it makes
	// all of them fail at cache creation, so each group produces only a cache error.
	runErroredSuites := func(failfast bool) uint {
		chartDir := copyChartToTemp(t, internalTestBasicChart)
		blockSnapshotDir(t, filepath.Join(chartDir, "tests", "__snapshot__"))

		buffer := new(bytes.Buffer)
		tr := &TestRunner{
			Printer:    printer.NewPrinter(buffer, nil),
			TestFiles:  []string{"tests/*_test.yaml"},
			Parallel:   true,
			Failfast:   failfast,
			MaxWorkers: 1,
		}
		assert.False(t, tr.RunV3([]string{chartDir}), buffer.String())
		return tr.suiteCounting.failed
	}

	withoutFailfast := runErroredSuites(false)
	withFailfast := runErroredSuites(true)

	assert.Greater(t, withoutFailfast, uint(1), "fixture must have multiple suite groups")
	assert.Equal(t, withoutFailfast, withFailfast,
		"cache-creation errors must not make failfast skip later groups")
}
