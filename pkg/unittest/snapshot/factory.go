package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

const snapshotDirName = "__snapshot__"
const snapshotFileExt = ".snap"

// CreateSnapshotOfSuite retruns snapshot.Cache for suite file, create `__snapshot__` dir if not existed
func CreateSnapshotOfSuite(path string, isUpdating bool) (*Cache, error) {
	cacheDir := filepath.Join(filepath.Dir(path), snapshotDirName)
	if err := ensureDir(cacheDir); err != nil {
		return nil, err
	}
	cacheFileName := filepath.Base(path) + snapshotFileExt
	cache := &Cache{
		Filepath:   filepath.Join(cacheDir, cacheFileName),
		IsUpdating: isUpdating,
	}

	if err := cache.RestoreFromFile(); err != nil {
		return nil, err
	}
	return cache, nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0755)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("snapshot cache dir %s is not a directory", path)
	}
	return nil
}
