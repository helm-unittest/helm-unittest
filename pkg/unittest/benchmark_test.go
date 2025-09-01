package unittest_test

import (
	"fmt"
	"io"
	"os"
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
	files, err := copyFile("testdata/chart-benchmark/tests", "main_test.yaml", 400)
	assert.NoError(b, err)

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

	for b.Loop() {
		_ = runner.RunV3([]string{"testdata/chart-benchmark"})
	}
}

func copyFile(dir, src string, times int) ([]string, error) {
	// Open the source file for reading
	in, err := os.ReadFile(fmt.Sprintf("%s/%s", dir, src))
	if err != nil {
		return nil, err
	}
	var result []string

	for i := 0; i < times; i++ {
		fileName := fmt.Sprintf("%s/%d-%s", dir, i, src)
		err := os.WriteFile(fileName, in, 0644)
		if err != nil {
			return nil, err
		}
		result = append(result, fileName)
	}

	return result, nil
}
