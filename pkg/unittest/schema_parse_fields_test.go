package unittest_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertionsSupportingParse lists every assertion whose schema block must
// declare the parse and innerPath properties. It mirrors the set of validators
// embedding ParseOptions.
//
// isNull and isNotNull also support parse at runtime but have no schema blocks
// at all, a pre-existing gap in the schema that is out of scope here.
var assertionsSupportingParse = []string{
	"equal", "notEqual",
	"isSubset", "isNotSubset",
	"contains", "notContains",
	"exists", "notExists",
	"isNullOrEmpty", "isNotNullOrEmpty",
	"isEmpty", "isNotEmpty",
	"isType", "isNotType",
	"lengthEqual", "notLengthEqual",
	"greaterOrEqual", "notGreaterOrEqual",
	"lessOrEqual", "notLessOrEqual",
}

// assertionsRejectingParse lists assertions that must NOT declare the parse
// fields, so IDE validation matches the runtime rejection.
var assertionsRejectingParse = []string{
	"matchRegex", "notMatchRegex",
	"matchRegexRaw", "notMatchRegexRaw",
	"equalRaw", "notEqualRaw",
	"isKind", "isAPIVersion",
	"hasDocuments", "containsDocument",
	"failedTemplate", "notFailedTemplate",
	"matchSnapshot", "matchSnapshotRaw",
}

func loadTestSuiteSchema(t *testing.T) map[string]any {
	t.Helper()

	raw, err := os.ReadFile("../../schema/helm-testsuite.json")
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(raw, &schema))

	return schema
}

// findAssertionProperties locates the property map for a named assertion within
// the schema's asserts oneOf list.
func findAssertionProperties(t *testing.T, schema map[string]any, name string) (map[string]any, bool) {
	t.Helper()

	tests, ok := schema["properties"].(map[string]any)["tests"].(map[string]any)
	require.True(t, ok)
	items, ok := tests["items"].(map[string]any)
	require.True(t, ok)
	asserts, ok := items["properties"].(map[string]any)["asserts"].(map[string]any)
	require.True(t, ok)
	assertItems, ok := asserts["items"].(map[string]any)
	require.True(t, ok)
	oneOf, ok := assertItems["oneOf"].([]any)
	require.True(t, ok)

	for _, entry := range oneOf {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		properties, ok := entryMap["properties"].(map[string]any)
		if !ok {
			continue
		}
		definition, ok := properties[name].(map[string]any)
		if !ok {
			continue
		}
		assertionProperties, ok := definition["properties"].(map[string]any)
		if !ok {
			return nil, false
		}
		return assertionProperties, true
	}

	return nil, false
}

func TestSchemaDeclaresParseFieldsForSupportedAssertions(t *testing.T) {
	schema := loadTestSuiteSchema(t)

	for _, name := range assertionsSupportingParse {
		t.Run(name, func(t *testing.T) {
			properties, found := findAssertionProperties(t, schema, name)
			require.True(t, found, "no schema block found for assertion %s", name)

			parseProperty, ok := properties["parse"].(map[string]any)
			require.True(t, ok, "assertion %s must declare a parse property", name)
			assert.Equal(t, "string", parseProperty["type"])
			assert.Equal(t, []any{"json", "yaml"}, parseProperty["enum"])
			assert.NotEmpty(t, parseProperty["description"])
			assert.NotEmpty(t, parseProperty["markdownDescription"])

			innerPathProperty, ok := properties["innerPath"].(map[string]any)
			require.True(t, ok, "assertion %s must declare an innerPath property", name)
			assert.Equal(t, "string", innerPathProperty["type"])
			assert.NotEmpty(t, innerPathProperty["description"])
			assert.NotEmpty(t, innerPathProperty["markdownDescription"])
		})
	}
}

func TestSchemaOmitsParseFieldsForUnsupportedAssertions(t *testing.T) {
	schema := loadTestSuiteSchema(t)

	for _, name := range assertionsRejectingParse {
		t.Run(name, func(t *testing.T) {
			properties, found := findAssertionProperties(t, schema, name)
			if !found {
				t.Skipf("assertion %s has no schema block with properties", name)
			}

			assert.NotContains(t, properties, "parse",
				"assertion %s must not declare parse; the runtime rejects it", name)
			assert.NotContains(t, properties, "innerPath",
				"assertion %s must not declare innerPath; the runtime rejects it", name)
		})
	}
}
