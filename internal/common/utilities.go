package common

import (
	"bytes"
	"strings"
	"testing"

	"io"

	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"
	yaml "sigs.k8s.io/yaml"
)

type YamlNode struct {
	Node yamlv3.Node
}

func NewYamlNode() YamlNode {
	return YamlNode{
		Node: yamlv3.Node{},
	}
}

// YamlNewDecoder returns a new decoder that reads from r.
func YamlNewDecoder(r io.Reader) *yamlv3.Decoder {
	return yamlv3.NewDecoder(r)
}

// YamlNewEncoder returns a new encoder that writes to w.
func YamlNewEncoder(w io.Writer) *yamlv3.Encoder {
	return yamlv3.NewEncoder(w)
}

// TrustedMarshalYAML marshal yaml without error returned, if an error happens it panics
func TrustedMarshalYAML(d interface{}) string {
	byteBuffer := new(bytes.Buffer)
	yamlEncoder := yamlv3.NewEncoder(byteBuffer)
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
	yamlDecoder := yamlv3.NewDecoder(strings.NewReader(d))
	if err := yamlDecoder.Decode(&parsedYaml); err != nil {
		panic(err)
	}
	return parsedYaml
}

func YamlToJson(in string) ([]byte, error) {
	return yaml.YAMLToJSON([]byte(in))
}

func YmlUnmarshal(in string, out interface{}) error {
	return yamlv3.Unmarshal([]byte(in), out)
}

func YmlMarshall(in interface{}) (string, error) {
	out, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func YmlUnmarshalTestHelper(input string, out any, t *testing.T) {
	t.Helper()
	err := YmlUnmarshal(input, out)
	assert.NoError(t, err)
}

func YmlMarshallTestHelper(in interface{}, t *testing.T) string {
	t.Helper()
	out, err := yaml.Marshal(in)
	assert.NoError(t, err)
	return string(out)
}

func SplitBefore(s, sep string) []string {
	var out []string

	// this can be omitted if staying analogous with SplitAfter is not a requirement
	if strings.HasPrefix(s, sep) || s == "" {
		out = append(out, "")
	}

	for len(s) > 0 {
		i := strings.Index(s[1:], sep)
		if i == -1 {
			out = append(out, s)
			break
		}

		out = append(out, s[:i+1])
		s = s[i+1:]
	}
	return out
}
