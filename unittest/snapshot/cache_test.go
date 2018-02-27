package snapshot_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestCacheWhenFirstTime(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test")
	cache := Cache{Filepath: filepath.Join(dir, "cache_test.yaml")}
	err := cache.RestoreFromFile()

	a := assert.New(t)
	a.Nil(err)
	a.False(cache.Existed)
	a.False(cache.Changed())

	cache.Compare("new test", 1, map[string]interface{}{
		"a": map[string]string{
			"b": "c",
		},
	})
	a.True(cache.Changed())

	storeErr := cache.StoreToFile()
	a.Nil(storeErr)
	a.True(cache.Existed)

	expectedCacheContent := `new test:
  1: |
    a:
      b: c
`
	bytes, _ := ioutil.ReadFile(cache.Filepath)
	a.Equal(expectedCacheContent, string(bytes))

	os.RemoveAll(dir)
}

func TestCacheWhenCachedBefore(t *testing.T) {
	lastTimeContent := `cached before:
  1: |
    a:
      b: c

`

	dir, _ := ioutil.TempDir("", "test")
	cacheFile := filepath.Join(dir, "cache_test.yaml")
	ioutil.WriteFile(cacheFile, []byte(lastTimeContent), os.ModePerm)

	cache := Cache{Filepath: cacheFile}
	err := cache.RestoreFromFile()

	a := assert.New(t)
	a.Nil(err)
	a.True(cache.Existed)
	a.True(cache.Changed())

	cache.Compare("cached before", 1, map[string]interface{}{
		"a": map[string]string{
			"b": "c",
		},
	})
	a.False(cache.Changed())

	cache.Compare("cached before", 1, map[string]interface{}{
		"x": map[string]string{
			"y": "z",
		},
	})
	a.True(cache.Changed())

	storeErr := cache.StoreToFile()
	a.Nil(storeErr)
	a.True(cache.Existed)

	expectedCacheContent := `cached before:
  1: |
    x:
      "y": z
`
	bytes, _ := ioutil.ReadFile(cache.Filepath)
	a.Equal(expectedCacheContent, string(bytes))

	os.RemoveAll(dir)
}
