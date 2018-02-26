package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
)

const snapshotDirName = "__snapshot__"
const snapshotFileExt = ".snap"

func CreateSnapshotOfFile(path string) (*SnapshotCache, error) {
	cacheDir := filepath.Join(filepath.Dir(path), snapshotDirName)
	if err := ensureDir(cacheDir); err != nil {
		return nil, err
	}
	cacheFileName := filepath.Base(path) + snapshotFileExt
	cache := &SnapshotCache{
		Filepath: filepath.Join(cacheDir, cacheFileName),
	}

	if err := cache.RestoreFromFile(); err != nil {
		return nil, err
	}
	return cache, nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, os.ModePerm)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("snapshot cache dir %s is not a directory", path)
	}
	return nil
}
