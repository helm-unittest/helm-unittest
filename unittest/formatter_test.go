package unittest_test

import (
    "os"
    "path/filepath"
    "testing"

    . "github.com/lrills/helm-unittest/unittest"
    "github.com/stretchr/testify/assert"
)

func TestNewFormatterWithEmptyOutputFile(t *testing.T) {
    given := ""
    sut := NewFormatter(given)
    assert.Nil(t, sut)
}

func TestNewFormatterWithOutputFile(t *testing.T) {
    assert := assert.New(t)
    given := "../__fixtures__/test-output/test_output.xml"
    givenDirectory := filepath.Dir(given)
    defer os.Remove(givenDirectory)
    sut := NewFormatter(given)
    assert.NotNil(sut)
    assert.DirExists(givenDirectory)
}
