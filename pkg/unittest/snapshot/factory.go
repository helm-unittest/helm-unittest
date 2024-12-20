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
	fmt.Println("ensureDir")
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0755)
		}
		fmt.Println("ensureDir err, ", err)
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("snapshot cache dir %s is not a directory", path)
	}
	return nil
}
