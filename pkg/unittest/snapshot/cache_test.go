package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

const cache_before string = "cached before"

var lastTimeContent = `cached before:
  1: |
    a:
      b: c
  2: |
    d:
      e: f
`

var snapshot1 = "a:\n  b: c\n"
var content1 = map[string]interface{}{
	"a": map[string]string{
		"b": "c",
	},
}

var snapshot2 = "d:\n  e: f\n"
var content2 = map[string]interface{}{
	"d": map[string]string{
		"e": "f",
	},
}

var snapshotNew = "x:\n  \"y\": z\n"
var contentNew = map[string]interface{}{
	"x": map[string]string{
		"y": "z",
	},
}

var contentForRegex = map[string]interface{}{
	"x": map[string]string{
		"y": "abrakadabra",
	},
}

func createCache(assert *assert.Assertions, existed bool) *Cache {
	dir, _ := os.MkdirTemp("", "test")
	cacheFile := filepath.Join(dir, "cache_test.yaml")
	if existed {
		err := os.WriteFile(cacheFile, []byte(lastTimeContent), os.ModePerm)
		if err != nil {
			assert.FailNow("Failed to create cache file")
		}
	}

	return &Cache{Filepath: cacheFile}
}

func createCacheResult(index uint, passed bool, cachedSnapshot, newSnapshot string) *CompareResult {
	return &CompareResult{
		Test:           cache_before,
		Index:          index,
		Passed:         passed,
		CachedSnapshot: cachedSnapshot,
		NewSnapshot:    newSnapshot,
	}
}

func verifyCache(assert *assert.Assertions, cache *Cache, exists, changed bool, current, inserted, updated, failed, vanished uint) {
	assert.Equal(exists, cache.Existed)
	assert.Equal(changed, cache.Changed())
	assert.Equal(current, cache.CurrentCount())
	assert.Equal(inserted, cache.InsertedCount())
	assert.Equal(updated, cache.UpdatedCount())
	assert.Equal(failed, cache.FailedCount())
	assert.Equal(vanished, cache.VanishedCount())
}

func TestCacheWhenFirstTime(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, false)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, false, false, 0, 0, 0, 0, 0)

	cache.Compare("new test", 1, content1)
	verifyCache(a, cache, false, true, 1, 1, 0, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 1, 1, 0, 0, 0)

	expectedCacheContent := `new test:
  1: |
    a:
      b: c
`
	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(expectedCacheContent, string(bytes))
}

func TestCacheWhenNotChanged(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	result := cache.Compare(cache_before, 1, content1)
	a.Equal(createCacheResult(1, true, snapshot1, snapshot1), result)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, content2)
	a.Equal(createCacheResult(2, true, snapshot2, snapshot2), result2)
	verifyCache(a, cache, true, false, 2, 0, 0, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.False(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, false, 2, 0, 0, 0, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(lastTimeContent, string(bytes))
}

func TestCacheWhenChanged(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, contentNew)
	a.Equal(createCacheResult(2, false, snapshot2, snapshotNew), result2)
	verifyCache(a, cache, true, true, 2, 0, 1, 1, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.False(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 2, 0, 1, 1, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(lastTimeContent, string(bytes))
}

func TestCacheWhenNotChangedIfIsUpdating(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	cache.IsUpdating = true
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	result := cache.Compare(cache_before, 1, content1)
	a.Equal(createCacheResult(1, true, snapshot1, snapshot1), result)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, content2)
	a.Equal(createCacheResult(2, true, snapshot2, snapshot2), result2)
	verifyCache(a, cache, true, false, 2, 0, 0, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.False(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, false, 2, 0, 0, 0, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(lastTimeContent, string(bytes))
}

func TestCacheWhenChangedIfIsUpdating(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	cache.IsUpdating = true
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, contentNew)
	a.Equal(createCacheResult(2, true, snapshot2, snapshotNew), result2)
	verifyCache(a, cache, true, true, 2, 0, 1, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 2, 0, 1, 0, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(`cached before:
  1: |
    a:
      b: c
  2: |
    x:
      "y": z
`, string(bytes))
}

func TestCacheWhenHasVanished(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(`cached before:
  1: |
    a:
      b: c
`, string(bytes))
}

func TestCacheWhenHasInserted(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	cache.Compare(cache_before, 2, content2)
	verifyCache(a, cache, true, false, 2, 0, 0, 0, 0)

	result3 := cache.Compare(cache_before, 3, contentNew)
	a.Equal(createCacheResult(3, true, "", snapshotNew), result3)
	verifyCache(a, cache, true, true, 3, 1, 0, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 3, 1, 0, 0, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(`cached before:
  1: |
    a:
      b: c
  2: |
    d:
      e: f
  3: |
    x:
      "y": z
`, string(bytes))
}

func TestCacheWhenNewOneAtMiddle(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, contentNew)
	a.Equal(createCacheResult(2, false, snapshot2, snapshotNew), result2)
	verifyCache(a, cache, true, true, 2, 0, 1, 1, 0)

	result3 := cache.Compare("cached before", 3, content2)
	a.Equal(createCacheResult(3, true, "", snapshot2), result3)
	verifyCache(a, cache, true, true, 3, 1, 1, 1, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 3, 1, 1, 1, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(`cached before:
  1: |
    a:
      b: c
  2: |
    d:
      e: f
  3: |
    d:
      e: f
`, string(bytes))
}

func TestCacheWhenNewOneAtMiddleIfIsUpdating(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	cache.IsUpdating = true
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	cache.Compare(cache_before, 1, content1)
	verifyCache(a, cache, true, true, 1, 0, 0, 0, 1)

	result2 := cache.Compare(cache_before, 2, contentNew)
	a.Equal(createCacheResult(2, true, snapshot2, snapshotNew), result2)
	verifyCache(a, cache, true, true, 2, 0, 1, 0, 0)

	result3 := cache.Compare(cache_before, 3, content2)
	a.Equal(createCacheResult(3, true, "", snapshot2), result3)
	verifyCache(a, cache, true, true, 3, 1, 1, 0, 0)

	stored, storeErr := cache.StoreToFileIfNeeded()
	a.True(stored)
	a.Nil(storeErr)
	verifyCache(a, cache, true, true, 3, 1, 1, 0, 0)

	bytes, _ := os.ReadFile(cache.Filepath)
	a.Equal(`cached before:
  1: |
    a:
      b: c
  2: |
    x:
      "y": z
  3: |
    d:
      e: f
`, string(bytes))
}

func TestCacheWhenRegexMatch(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)

	result := cache.Compare(cache_before, 1, content1, WithMatchRegexPattern("b: c"))
	a.True(result.Passed)

	result = cache.Compare(cache_before, 1, content1, WithMatchRegexPattern("bcd"))
	a.False(result.Passed)
	a.Contains(result.Msg, "pattern 'bcd' not found in snapshot")
}

func TestCacheWhenNotRegexMatch(t *testing.T) {
	a := assert.New(t)
	cache := createCache(a, true)
	err := cache.RestoreFromFile()

	a.Nil(err)
	verifyCache(a, cache, true, true, 0, 0, 0, 0, 2)

	result := cache.Compare(cache_before, 1, content1, WithNotMatchRegexPattern("not-exist"))
	a.True(result.Passed)

	result = cache.Compare(cache_before, 1, content1, WithNotMatchRegexPattern("b: c"))
	a.False(result.Passed)
	a.Contains(result.Msg, "pattern 'b: c' should not be in snapshot")
}
