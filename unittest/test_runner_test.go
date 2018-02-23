package unittest_test

import (
	"os"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/stretchr/testify/assert"
)

func TestTestRunner(t *testing.T) {
	runner := TestRunner{ChartsPath: []string{"../__fixtures__/basic"}}
	passed := runner.Run(&Printer{Writer: os.Stdout, Colored: true}, TestConfig{})
	assert.True(t, passed)
}
