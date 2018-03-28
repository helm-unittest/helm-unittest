package unittest_test

import (
	"bytes"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

func TestRunnerOkWithPassedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, buffer.String())
}

func TestRunnerOkWithFailedTests(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			TestFiles: []string{"tests_failed/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.False(t, passed)
	cupaloy.SnapshotT(t, buffer.String())
}

func TestRunnerWithTestsInSubchart(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: true,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, buffer.String())
}

func TestRunnerWithTestsInSubchartButFlagFalse(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer: NewPrinter(buffer, nil),
		Config: TestConfig{
			WithSubChart: false,
			TestFiles:    []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/with-subchart"})
	assert.True(t, passed)
	cupaloy.SnapshotT(t, buffer.String())
}
