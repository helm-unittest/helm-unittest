package common

import (
	"bytes"

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
