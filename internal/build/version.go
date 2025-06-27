package build

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	log "github.com/sirupsen/logrus"
)

// the plugin version is injected by linker at build time
var version string = "0.1.0"

// GetVersion returns the version of the plugin
func GetVersion() string {
	// in compiled binary, version should be set
	log.WithField(common.LOG_BUILD_VERSION, "getversion").Debugln("build-version", version)
	return version
}
