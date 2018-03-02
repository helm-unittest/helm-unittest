package unittest_test

import (
	"os"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

func TestTestRunner(t *testing.T) {
	runner := TestRunner{
		Logger: &Printer{Writer: os.Stdout, Colored: true},
		Config: TestConfig{
			TestFiles: []string{"tests/*_test.yaml"},
		},
	}
	passed := runner.Run([]string{"../__fixtures__/basic"})
	assert.True(t, passed)
}
