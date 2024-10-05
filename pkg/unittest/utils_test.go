package unittest_test

import (
	"testing"

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

func TestIsYaml(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"yaml file", "test.yaml", true},
		{"yml file", "config.yml", true},
		{"template file", "template.tpl", true},
		{"text file", "data.txt", false},
		{"no extension", "file", false},
		{"unknown extension", "unknown.ext", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsYaml(tc.fileName)
			if got != tc.want {
				t.Errorf("IsYaml(%q) = %v, want %v", tc.fileName, got, tc.want)
			}
		})
	}
}
