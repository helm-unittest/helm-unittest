package yamlutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

func YmlUnmarshalMap(input string, t *testing.T) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(input), &data)
	assert.NoError(t, err)
	return data
}

func YmlMarshal(input interface{}, t *testing.T) string {
	t.Helper()
	data, err := yaml.Marshal(&input)
	assert.NoError(t, err)
	return string(data)
}
