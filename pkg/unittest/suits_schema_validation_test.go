package unittest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

const testSuiteSchemaLocal string = "../../schema/helm-testsuite.json"

type schemaValidation struct {
	FilePath     string
	FileContents []string
}

func readAllFiles(dirPath string) ([]schemaValidation, error) {
	var result []schemaValidation

	// Read the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	// Iterate through each entry
	for _, entry := range entries {
		// Skip if it's a directory
		if entry.IsDir() {
			continue
		}
		// Construct full file path
		filePath := filepath.Join(dirPath, entry.Name())
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}
		// Append content to slice
		var fileContents []string
		fileContents = append(fileContents, string(content))
		result = append(result, schemaValidation{
			FilePath: filePath,
			FileContents: fileContents,
		})
	}

	return result, nil
}

// TestValidateSuitsAgainstSchema validates Helm test suites against a defined JSON schema.
//
// It performs the following steps:
// 1. Loads the Helm test suite schema from a specified JSON file.
// 2. Reads multiple Helm test suite files from a designated directory.
// 3. Converts YAML content of each test suite to JSON format.
// 4. Validates each JSON-converted test suite against the loaded schema.
// 5. Asserts that all test suites adhere to the schema and provides detailed error messages for validation failures.
//
// Overall, this test ensures the consistency and correctness of Helm test suites by verifying their compliance with the defined schema.
func TestValidateExampleChartsWithTestSuitsAgainstLocalSchema(t *testing.T) {
	fullPath, _ := filepath.Abs(testSuiteSchemaLocal)

	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file:///%s", fullPath))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	assert.NoError(t, err)
	assert.NotEmpty(t, schema, fmt.Sprintf("Schema '%s' is not valid!!!", fullPath))

	tests := []struct {
		testsPath    string
	}{
		{
			testsPath: "../../test/data/v3/basic/tests",
		},
		{
			testsPath: "../../test/data/v3/full-snapshot/tests",
		},
		{
			testsPath: "../../test/data/v3/global-double-setting/tests",
		},
		{
			testsPath: "../../test/data/v3/nested_glob/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subsubcharts/tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testsPath, func(t *testing.T) {
			content, err := readAllFiles(tt.testsPath)
			assert.NoError(t, err)
			for _, el := range content {
				for _, content := range el.FileContents {
					json, err := yaml.YAMLToJSON([]byte(content))

					assert.NoError(t, err)
					assert.NotEmpty(t, json)

					loader := gojsonschema.NewStringLoader(string(json))
					assert.NotEmpty(t, loader)
					assert.NoError(t, err)

					result, err := gojsonschema.Validate(schemaLoader, loader)
					assert.NoError(t, err)

					assert.True(t, result.Valid(), fmt.Sprintf("Schema '%s' and the document '%s' is not valid!!!", fullPath, el.FilePath))

					if !result.Valid() {
						fmt.Printf("See errors:\n")
						for _, desc := range result.Errors() {
							fmt.Printf("- %s\n", desc)
						}
					}
				}
			}

		})
	}
}
