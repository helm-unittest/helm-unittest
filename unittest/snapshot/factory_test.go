package snapshot_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshotOfFileWhenNoCacheDir(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test")
	cache, err := CreateSnapshotOfFile(filepath.Join(dir, "service_test.yaml"))

	a := assert.New(t)
	a.Nil(err)
	a.Equal(filepath.Join(dir, "__snapshot__", "service_test.yaml.snap"), cache.Filepath)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))
	a.Nil(err)
	a.True(info.IsDir())

	_, statCacheFileErr := os.Stat(cache.Filepath)
	a.False(cache.Existed)
	a.True(os.IsNotExist(statCacheFileErr))

	os.RemoveAll(dir)
}

func TestCreateSnapshotOfFileWhenCacheDirExisted(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test")
	os.Mkdir(filepath.Join(dir, "__snapshot__"), os.ModePerm)
	cache, err := CreateSnapshotOfFile(filepath.Join(dir, "service_test.yaml"))

	a := assert.New(t)
	a.Nil(err)
	a.Equal(filepath.Join(dir, "__snapshot__", "service_test.yaml.snap"), cache.Filepath)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))
	a.Nil(err)
	a.True(info.IsDir())

	_, statCacheFileErr := os.Stat(cache.Filepath)
	a.False(cache.Existed)
	a.True(os.IsNotExist(statCacheFileErr))

	os.RemoveAll(dir)
}

func TestCreateSnapshotOfFileWhenCacheFileExisted(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test")
	os.Mkdir(filepath.Join(dir, "__snapshot__"), os.ModePerm)
	err := ioutil.WriteFile(filepath.Join(dir, "__snapshot__", "service_test.yaml.snap"), []byte(`a test:
  1: |
    a:
      b: c
`), os.ModePerm)

	cache, err := CreateSnapshotOfFile(filepath.Join(dir, "service_test.yaml"))

	a := assert.New(t)
	a.Nil(err)
	a.Equal(filepath.Join(dir, "__snapshot__", "service_test.yaml.snap"), cache.Filepath)

	info, err := os.Stat(filepath.Join(dir, "__snapshot__"))
	a.Nil(err)
	a.True(info.IsDir())

	info, statCacheFileErr := os.Stat(cache.Filepath)
	a.True(cache.Existed)
	a.Nil(statCacheFileErr)

	os.RemoveAll(dir)
}
