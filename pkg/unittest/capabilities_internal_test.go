package unittest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
)

// capabilitiesV4 must not mutate the shared global chartcommon.DefaultCapabilities
// pointer; it should return a copy with the job-specific overrides applied.
func TestCapabilitiesV4DoesNotMutateGlobal(t *testing.T) {
	originalMajor := chartcommon.DefaultCapabilities.KubeVersion.Major
	originalMinor := chartcommon.DefaultCapabilities.KubeVersion.Minor
	originalAPIVersions := chartcommon.DefaultCapabilities.APIVersions

	job := &TestJob{}
	job.Capabilities.MajorVersion = "1"
	job.Capabilities.MinorVersion = "42"
	job.Capabilities.APIVersions = []string{"example.com/v1"}

	caps := job.capabilitiesV4()

	assert.Equal(t, "42", caps.KubeVersion.Minor)
	assert.Equal(t, chartcommon.VersionSet([]string{"example.com/v1"}), caps.APIVersions)

	assert.Equal(t, originalMajor, chartcommon.DefaultCapabilities.KubeVersion.Major)
	assert.Equal(t, originalMinor, chartcommon.DefaultCapabilities.KubeVersion.Minor)
	assert.Equal(t, originalAPIVersions, chartcommon.DefaultCapabilities.APIVersions)
}
