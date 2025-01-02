package unittest

import (
	"bytes"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/printer"
	"github.com/stretchr/testify/assert"
)

func TestV3RunnerWith_Fixture_Chart_ErrorWhenMetaCharacters(t *testing.T) {
	buffer := new(bytes.Buffer)
	runner := TestRunner{
		Printer:   printer.NewPrinter(buffer, nil),
		TestFiles: []string{"tests/*_test.yaml"},
	}
	passed := runner.RunV3([]string{"testdata/chart01"})
	assert.True(t, passed, buffer.String())
}
