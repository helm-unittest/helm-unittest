package helmutils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Package motivation https://github.com/helm/helm/blob/663a896f4a815053445eec4153677ddc24a0a361/pkg/chart/loader/directory.go

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

// DirLoader loads a chart from a directory
type DirLoader struct {
	path  string
	rules Rules
}

// Load loads the chart
func (l DirLoader) Load() (*chart.Chart, error) {
	return LoadDir(l.path, l.rules)
}

// LoadDir loads from a directory.
//
// This loads charts only from directories.
func LoadDir(dir string, rules Rules) (*chart.Chart, error) {
	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	var files []*loader.BufferedFile
	topdir += string(filepath.Separator)

	walk := func(name string, fi os.FileInfo, err error) error {
		n := strings.TrimPrefix(name, topdir)
		if n == "" {
			return nil
		}

		// Normalize to / since it will also work on Windows
		n = filepath.ToSlash(n)

		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(n, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If rules matches, skip this file.
		if rules.Ignore(n, fi) {
			return nil
		}

		// Irregular files include devices, sockets, and other uses of files that
		// are not regular files. In Go they have a file mode type bit set.
		// See https://golang.org/pkg/os/#FileMode for examples.
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", name)
		}

		if fi.Size() > loader.MaxDecompressedFileSize {
			return fmt.Errorf("chart file %q is larger than the maximum file size %d", fi.Name(), loader.MaxDecompressedFileSize)
		}

		data, err := os.ReadFile(name)
		if err != nil {
			return errors.Wrapf(err, "error reading %s", n)
		}

		data = bytes.TrimPrefix(data, utf8bom)

		files = append(files, &loader.BufferedFile{Name: n, Data: data})
		return nil
	}
	if err = Walk(topdir, walk); err != nil {
		return nil, err
	}

	return loader.LoadFiles(files)
}
