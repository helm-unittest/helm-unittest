package unittest_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/stretchr/testify/assert"

	"github.com/xeipuuv/gojsonschema"
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
			FilePath:     filePath,
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
		testsPath string
	}{
		{
			testsPath: "../../test/data/v3/basic/tests",
		},
		{
			testsPath: "../../test/data/v3/basic/tests_failed",
		},
		{
			testsPath: "../../test/data/v3/failing-template/tests",
		},
		{
			testsPath: "../../test/data/v3/full-snapshot/tests",
		},
		{
			testsPath: "../../test/data/v3/global-double-setting/tests",
		},
		{
			testsPath: "../../test/data/v3/library-chart/tests/chart/tests/unit",
		},
		{
			testsPath: "../../test/data/v3/nested_glob/tests",
		},
		{
			testsPath: "../../test/data/v3/with-document-select/tests",
		},
		{
			testsPath: "../../test/data/v3/with-document-select/tests_failed",
		},
		{
			testsPath: "../../test/data/v3/with-files/tests",
		},
		{
			testsPath: "../../test/data/v3/with-k8s-fake-client/tests",
		},
		{
			testsPath: "../../test/data/v3/with-post-renderer/tests",
		},
		{
			testsPath: "../../test/data/v3/with-samenamesubsubcharts/tests",
		},
		{
			testsPath: "../../test/data/v3/with-samenamesubsubcharts/charts/with-subsubchartssub/tests",
		},
		{
			testsPath: "../../test/data/v3/with-samenamesubsubcharts/charts/with-subsubchartssub/charts/with-subsubchartssubsub/tests",
		},
		{
			testsPath: "../../test/data/v3/with-schema/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subchart/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subchart/charts/child-chart/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subfolder/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subsubcharts/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subsubcharts/charts/with-subsubchartssub/tests",
		},
		{
			testsPath: "../../test/data/v3/with-subsubcharts/charts/with-subsubchartssub/charts/with-subsubchartssubsub/tests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testsPath, func(t *testing.T) {
			content, err := readAllFiles(tt.testsPath)
			assert.NoError(t, err)
			for _, el := range content {
				for _, content := range el.FileContents {
					json, err := common.YamlToJson(content)

					assert.NoError(t, err)
					assert.NotEmpty(t, json)

					loader := gojsonschema.NewStringLoader(string(json))
					assert.NotEmpty(t, loader)
					assert.NoError(t, err)

					result, err := gojsonschema.Validate(schemaLoader, loader)
					assert.NoError(t, err)

					assert.True(t, result.Valid(), fmt.Sprintf("Schema '%s' and the document '%s' is not valid!!!", fullPath, el.FilePath))

					if !result.Valid() {
						fmt.Printf("See errors for file%s:\n", el.FilePath)
						for _, desc := range result.Errors() {
							fmt.Printf("- %s\n", desc)
						}
					}
				}
			}

		})
	}
}
