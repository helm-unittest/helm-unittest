package unittest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v3util "helm.sh/helm/v3/pkg/chartutil"
)

// capabilitiesV3 must not mutate the shared global v3util.DefaultCapabilities
// pointer; it should return a copy with the job-specific overrides applied.
func TestCapabilitiesV3DoesNotMutateGlobal(t *testing.T) {
	originalMajor := v3util.DefaultCapabilities.KubeVersion.Major
	originalMinor := v3util.DefaultCapabilities.KubeVersion.Minor
	originalAPIVersions := v3util.DefaultCapabilities.APIVersions

	job := &TestJob{}
	job.Capabilities.MajorVersion = "1"
	job.Capabilities.MinorVersion = "42"
	job.Capabilities.APIVersions = []string{"example.com/v1"}

	caps := job.capabilitiesV3()

	assert.Equal(t, "42", caps.KubeVersion.Minor)
	assert.Equal(t, v3util.VersionSet([]string{"example.com/v1"}), caps.APIVersions)

	assert.Equal(t, originalMajor, v3util.DefaultCapabilities.KubeVersion.Major)
	assert.Equal(t, originalMinor, v3util.DefaultCapabilities.KubeVersion.Minor)
	assert.Equal(t, originalAPIVersions, v3util.DefaultCapabilities.APIVersions)
}
