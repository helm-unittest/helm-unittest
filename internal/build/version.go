package build

import (
	"os"
	"path/filepath"

	"github.com/helm-unittest/helm-unittest/internal/common"
)

// the plugin version is injected by linker at build time
var version string

// GetVersion returns the version of the plugin
func GetVersion() string {
	// in compiled binary, version should be set
	if version != "" {
		return version
	}

	// for development, we can read the version from plugin.yaml
	pluginVersion := readVersionFromPluginYaml()
	if pluginVersion != "" {
		return pluginVersion
	}

	return "unknown"
}

func readVersionFromPluginYaml() string {
	// Look for plugin.yaml at the project root
	pluginYamlPath := filepath.Join("..", "..", "plugin.yaml")

	content, err := os.ReadFile(pluginYamlPath)
	if err != nil {
		return ""
	}

	var plugin struct {
		Version string `yaml:"version"`
	}

	if err := common.YmlUnmarshal(string(content), &plugin); err != nil {
		return ""
	}

	return plugin.Version
}
