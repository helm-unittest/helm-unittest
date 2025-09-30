package unittest_test

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/printer"
	"github.com/stretchr/testify/assert"
)

// This benchmark test measures the CPU and memory performance of the RunV3 method in the
// TestRunner type when running against a large number of test files. It programmatically
// creates 400 copies of a sample test file, runs the test suite using these files,
// and cleans up the generated files after the benchmark completes.
// The test uses a silent printer to avoid output overhead during benchmarking.
func BenchmarkNewTestForCPUAndMemory(b *testing.B) {
	// Copy the sample test file 400 times to simulate a large number of test files
	const numCopies = 400
	const sampleTestFile = "testdata/chart-benchmark/tests/main_test.yaml"
	const testDir = "testdata/chart-benchmark/tests/"

	// Ensure the test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		err := os.MkdirAll(testDir, os.ModePerm)
		assert.NoError(b, err)
	}

	// Create copies of the sample test file
	files := make([]string, numCopies)
	for i := range numCopies {
		destFile := fmt.Sprintf("%s%d%s", testDir, i, "main_test.yaml")
		files[i] = destFile
		input, err := os.ReadFile(sampleTestFile)
		assert.NoError(b, err)
		err = os.WriteFile(destFile, input, 0644)
		assert.NoError(b, err)
	}

	b.Cleanup(func() {
		for _, file := range files {
			err := os.Remove(file)
			assert.NoError(b, err)
		}
	})

	runner := TestRunner{
		Printer:   printer.NewPrinter(io.Discard, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    true,
	}

	cpuProfileFile, cpuProfErr := os.Create("cpu_profile.prof")
	assert.NoError(b, cpuProfErr)
	defer func() {
		cerr := cpuProfileFile.Close()
		assert.NoError(b, cerr)
	}()
	profileFile, profErr := os.Create("mem_profile.prof")
	assert.NoError(b, profErr)
	defer func() {
		perr := profileFile.Close()
		assert.NoError(b, perr)
	}()

	runtime.GC() // get up-to-date statistics

	for b.Loop() {
		_ = runner.RunV3([]string{"testdata/chart-benchmark"})
	}

	pperr := pprof.StartCPUProfile(cpuProfileFile)
	assert.NoError(b, pperr)
	defer pprof.StopCPUProfile()
	hperr := pprof.WriteHeapProfile(profileFile)
	assert.NoError(b, hperr)
}
