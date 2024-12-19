package unittest_test

import (
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	ymlutils "github.com/helm-unittest/helm-unittest/pkg/unittest/yamlutils"
	"github.com/stretchr/testify/assert"
)

// unmarshalJobTestHelper unmarshall a YAML-encoded string into a TestJob struct.
// It extracts the majorVersion, minorVersion, and apiVersions fields from
// CapabilitiesFields and populates the corresponding fields in Capabilities.
// If apiVersions is nil, it sets APIVersions to nil. If it's a slice,
// it appends string values to APIVersions. Returns an error if unmarshaling
// or processing fails.
func unmarshalJobTestHelper(input string, out *TestJob, t *testing.T) {
	t.Helper()
	err := ymlutils.YmlUnmarshall(input, &out)
	assert.NoError(t, err)
	out.SetCapabilities()
}
