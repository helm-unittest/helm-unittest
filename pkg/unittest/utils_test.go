package unittest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
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
	err := common.YmlUnmarshal(input, &out)
	assert.NoError(t, err)
	out.SetCapabilities()
}

// writeToFile writes the provided string data to a file with the given filename.
// It returns an error if the file cannot be created or if there is an error during writing.
func writeToFile(data string, filename string) error {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}

	// Create the file with an absolute path
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
