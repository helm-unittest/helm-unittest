package unittest

import (
	"path/filepath"
)

func IsYaml(fileName string) bool {
	ext := filepath.Ext(fileName)
	validExtensions := []string{".yaml", ".yml", ".tpl"}

	for _, b := range validExtensions {
		if b == ext {
			return true
		}
	}

	return false
}
