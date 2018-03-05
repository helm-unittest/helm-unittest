package unittest_test

import (
	"os"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

func TestTestRunnerOkWithPassedTests(t *testing.T) {
	runner := TestRunner{
		Printer: NewPrinter(os.Stdout, nil),
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.True(t, passed)
}

func TestTestRunnerOkWithFailedTests(t *testing.T) {
	runner := TestRunner{
		Printer: NewPrinter(os.Stdout, nil),
		Config: TestConfig{
			TestFiles: []string{"tests_failed/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.False(t, passed)
}
