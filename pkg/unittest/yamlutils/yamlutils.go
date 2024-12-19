package yamlutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

func YmlUnmarshall(in string, out interface{}) error {
	err := yaml.Unmarshal([]byte(in), out)
	return err
}

func YmlUnmarshalTestHelper(input string, out any, t *testing.T) {
	t.Helper()
	err := YmlUnmarshall(input, out)
	assert.NoError(t, err)
}

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
