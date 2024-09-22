package unittest_test

import (
	"reflect"

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
	if val, ok := out.CapabilitiesFields["majorVersion"]; ok {
		out.Capabilities.MajorVersion = ConvertIToString(val)
	}
	if val, ok := out.CapabilitiesFields["minorVersion"]; ok {
		out.Capabilities.MinorVersion = ConvertIToString(val)
	}
	if val, ok := out.CapabilitiesFields["apiVersions"]; ok {
		if val == nil {
			// key capabilities.apiVersions key exist but is not set
			out.Capabilities.APIVersions = nil
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			for _, v := range val.([]interface{}) {
				if str, ok := v.(string); ok {
					out.Capabilities.APIVersions = append(out.Capabilities.APIVersions, str)
				}
			}
		}
	}
	return nil
}
