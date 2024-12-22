package common

import (
	"bytes"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

const (
	bsCode byte = byte('\\')
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

type YmlEscapeHandlers struct{}

// Escape function is required, as yaml library no longer maintained
// yaml unmaintained library issue https://github.com/go-yaml/yaml/pull/862
func (y *YmlEscapeHandlers) Escape(content string) []byte {
	if !strings.Contains(content, `\`) {
		return nil
	}
	return escapeBackslashes([]byte(content))
}

func escapeBackslashes(content []byte) []byte {
	var result bytes.Buffer
	for i := 0; i < len(content); i++ {
		if content[i] != bsCode {
			result.WriteByte(content[i])
			continue
		}
		count := 1
		for i+count < len(content) && content[i+count] == bsCode {
			count++
		}

		times := count
		if count%2 == 1 {
			times = count + 1
		}

		if count > 1 {
			i += count - 1
		}

		for j := 0; j < times; j++ {
			result.WriteByte('\\')
		}

	}
	return result.Bytes()
}
