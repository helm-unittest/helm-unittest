package common

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

// TrustedMarshalYAML marshal yaml without error returned, if an error happens it panics
func TrustedMarshalYAML(d interface{}) string {
	byteBuffer := new(bytes.Buffer)
	yamlEncoder := yaml.NewEncoder(byteBuffer)
	yamlEncoder.SetIndent(YAMLINDENTION)
	defer yamlEncoder.Close()
	if err := yamlEncoder.Encode(d); err != nil {
		panic(err)
	}
	return byteBuffer.String()
}

// TrustedUnmarshalYAML unmarshal yaml without error returned, if an error happens it panics
func TrustedUnmarshalYAML(d string) map[string]interface{} {
	parsedYaml := K8sManifest{}
	yamlDecoder := yaml.NewDecoder(strings.NewReader(d))
	if err := yamlDecoder.Decode(&parsedYaml); err != nil {
		panic(err)
	}
	return parsedYaml
}

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
