package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK8sManifestToString(t *testing.T) {
	manifest := K8sManifest{
		"app.kubernetes.io/version": "v1.0.0",
		"app.kubernetes.io/name":    "example",
	}

	expectedYAML := "app.kubernetes.io/name: example\napp.kubernetes.io/version: v1.0.0\n"

	result := manifest.ToString()

	assert.Equal(t, expectedYAML, result)
}
