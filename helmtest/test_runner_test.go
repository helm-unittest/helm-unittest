package helmtest_test

import (
	"os"
	"testing"

	. "github.com/lrills/helm-test/helmtest"
	"github.com/stretchr/testify/assert"
)

func TestTestRunner(t *testing.T) {
	runner := TestRunner{ChartsPath: []string{"../__fixtures__/basic"}}
	passed := runner.Run(&Printer{Writer: os.Stdout, Colored: true})
	assert.False(t, passed)
}
