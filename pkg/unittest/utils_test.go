package unittest_test

import (
	"gopkg.in/yaml.v3"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
)

// UnmarshalJob unmarshall a YAML-encoded string into a TestJob struct.
// It extracts the majorVersion, minorVersion, and apiVersions fields from
// CapabilitiesFields and populates the corresponding fields in Capabilities.
// If apiVersions is nil, it sets APIVersions to nil. If it's a slice,
// it appends string values to APIVersions. Returns an error if unmarshaling
// or processing fails.
func unmarshalJob(input string, out *TestJob) error {
	err := yaml.Unmarshal([]byte(input), &out)
	if err != nil {
		return err
	}
	out.SetCapabilities()
	return nil
}
