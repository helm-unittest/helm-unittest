package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshotOfSuiteReturnCacheRight(t *testing.T) {
	dir, _ := os.MkdirTemp("", "test")
	cache, err := CreateSnapshotOfSuite(filepath.Join(dir, "my_test.yaml"), true)
	cache2, err2 := CreateSnapshotOfSuite(filepath.Join(dir, "another_test.yaml"), false)

	a := assert.New(t)
	a.Nil(err)
	a.Equal(filepath.Join(dir, "__snapshot__", "my_test.yaml.snap"), cache.Filepath)
	a.True(cache.IsUpdating)

	a.Nil(err2)
	a.Equal(filepath.Join(dir, "__snapshot__", "another_test.yaml.snap"), cache2.Filepath)
	a.False(cache2.IsUpdating)
}

func TestCreateSnapshotOfSuiteWhenNoCacheDir(t *testing.T) {
	dir, _ := os.MkdirTemp("", "test")
	cache, _ := CreateSnapshotOfSuite(filepath.Join(dir, "service_test.yaml"), false)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))

	a := assert.New(t)
	a.Nil(err)
	a.True(info.IsDir())

	_, statCacheFileErr := os.Stat(cache.Filepath)
	a.False(cache.Existed)
	a.True(os.IsNotExist(statCacheFileErr))

	os.RemoveAll(dir)
}

func TestCreateSnapshotOfSuiteWhenCacheDirExisted(t *testing.T) {
	a := assert.New(t)

	dir, _ := os.MkdirTemp("", "test")
	dirErr := os.Mkdir(filepath.Join(dir, "__snapshot__"), os.ModePerm)
	if dirErr != nil {
		a.FailNow("Failed to create cache dir")
	}
	cache, _ := CreateSnapshotOfSuite(filepath.Join(dir, "service_test.yaml"), false)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))

	a.Nil(err)
	a.True(info.IsDir())

	_, statCacheFileErr := os.Stat(cache.Filepath)
	a.False(cache.Existed)
	a.True(os.IsNotExist(statCacheFileErr))

	os.RemoveAll(dir)
}

func TestCreateSnapshotOfSuiteWhenCacheFileExisted(t *testing.T) {
	a := assert.New(t)
	dir, _ := os.MkdirTemp("", "test")
	dirErr := os.Mkdir(filepath.Join(dir, "__snapshot__"), os.ModePerm)
	if dirErr != nil {
		a.FailNow("Failed to create cache dir")
	}
	fileErr := os.WriteFile(filepath.Join(dir, "__snapshot__", "service_test.yaml.snap"), []byte(`a test:
  1: |
    a:
      b: c
`), os.ModePerm)
	if fileErr != nil {
		a.FailNow("Failed to create cache file")
	}
	cache, _ := CreateSnapshotOfSuite(filepath.Join(dir, "service_test.yaml"), false)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))

	a.Nil(err)
	a.True(info.IsDir())

	_, statCacheFileErr := os.Stat(cache.Filepath)
	a.True(cache.Existed)
	a.Nil(statCacheFileErr)

	os.RemoveAll(dir)
}
