package unittest

import (
	"path/filepath"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/yargevad/filepathx"
)

const LOG_UTILS = "utils"

// GetFiles retrieves a list of files matching the given file patterns.
// If chartPath is provided, the patterns are treated as relative to the chartPath.
// If setAbsolute is true, the returned file paths are converted to absolute paths.
//
// chartPath is the base directory to search for files (can be empty).
// filePatterns is a slice of file patterns to match (e.g., "tests/*".yaml).
// setAbsolute indicates whether to return absolute file paths or relative paths.
//
// It returns a slice of file paths and an error if any occurred during processing.
func GetFiles(chartPath string, filePatterns []string, setAbsolute bool) ([]string, error) {
	log.WithField(LOG_UTILS, "get-files").Debugln("file-patterns:", filePatterns)
	var filesSet []string
	basePath := chartPath + "/" // Prepend chartPath with slash

	for _, pattern := range slices.Compact(filePatterns) {
		if filepath.IsAbs(pattern) {
			filesSet = append(filesSet, pattern) // Append absolute paths directly
		} else {
			var filePath string
			if strings.Contains(pattern, basePath) {
				filePath = pattern
			} else {
				filePath = filepath.Join(basePath, pattern)
			}
			files, err := filepathx.Glob(filePath)
			if err != nil {
				return nil, err
			}
			filesSet = append(filesSet, files...) // Append all files (relative)
		}
	}

	if setAbsolute {
		// If setAbsolute is true, convert the file paths to absolute paths
		for i, filePath := range filesSet {
			if !filepath.IsAbs(filePath) {
				absPath, _ := filepath.Abs(filePath)
				filesSet[i] = absPath
			}
		}
	}
	log.WithField(LOG_UTILS, "get-files").Debugln("chart-path:", chartPath, "fileset:", filesSet)
	return filesSet, nil
}
