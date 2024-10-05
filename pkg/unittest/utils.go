package unittest

import (
	"path/filepath"
)

func IsYaml(fileName string) bool {
	switch filepath.Ext(fileName) {
	case ".yaml", ".yml", ".tpl":
		return true
	default:
		return false
	}
}
