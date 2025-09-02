package unittest_test

import (
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
	runner := TestRunner{
		Printer:   printer.NewPrinter(io.Discard, nil),
		TestFiles: []string{"tests/*_test.yaml"},
		Strict:    true,
	}

	cpuProfileFile, cpuProfErr := os.Create("cpu_profile.prof")
	assert.NoError(b, cpuProfErr)
	defer cpuProfileFile.Close()
	profileFile, profErr := os.Create("mem_profile.prof")
	assert.NoError(b, profErr)
	defer profileFile.Close()

	runtime.GC() // get up-to-date statistics

	for b.Loop() {
		_ = runner.RunV3([]string{"testdata/chart-benchmark"})
	}

	pprof.StartCPUProfile(cpuProfileFile)
	defer pprof.StopCPUProfile()
	pprof.WriteHeapProfile(profileFile)
}
