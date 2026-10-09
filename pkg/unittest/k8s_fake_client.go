package unittest

import (
	"encoding/base64"
	"fmt"
	"maps"
	"path"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

type KubernetesFakeKindProps struct {
	ShouldErr  error                       `yaml:"should_err"`
	Gvr        schema.GroupVersionResource `yaml:"gvr"`
	Namespaced bool                        `yaml:"namespaced"`
}

type KubernetesFakeClientProvider struct {
	Scheme  map[string]KubernetesFakeKindProps `yaml:"scheme"`
	Objects []map[string]any                   `yaml:"objects"`
}

func (p *KubernetesFakeClientProvider) GetClientFor(apiVersion, kind string) (dynamic.NamespaceableResourceInterface, bool, error) {
	props := p.Scheme[path.Join(apiVersion, kind)]
	if props.ShouldErr != nil {
		return nil, false, props.ShouldErr
	}

	return fake.NewSimpleDynamicClient(runtime.NewScheme(), convertRuntimeObject(p.Objects)...).Resource(props.Gvr), props.Namespaced, nil
}

func convertRuntimeObject(input []map[string]any) []runtime.Object {
	result := make([]runtime.Object, len(input))

	for k, v := range input {
		result[k] = &unstructured.Unstructured{Object: normalizeSecret(v)}
	}

	return result
}

// normalizeSecret mimics what the api-server does when a v1/Secret is stored:
// the write-only stringData is base64 encoded into data and dropped.
func normalizeSecret(object map[string]any) map[string]any {
	if object["kind"] != "Secret" || object["apiVersion"] != "v1" {
		return object
	}

	rawStringData, found := object["stringData"]
	if !found {
		return object
	}
	stringData, ok := rawStringData.(map[string]any)
	if !ok && rawStringData != nil {
		return object // nothing we can encode, better than dropping data
	}

	data := map[string]any{}
	if existing, ok := object["data"].(map[string]any); ok {
		maps.Copy(data, existing)
	}
	for key, value := range stringData {
		data[key] = base64.StdEncoding.EncodeToString([]byte(stringify(value)))
	}

	// copy, the test definition is shared between test jobs
	normalized := maps.Clone(object)
	delete(normalized, "stringData")
	normalized["data"] = data

	return normalized
}

// stringify keeps non-string scalars such as `port: 8080` usable.
func stringify(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}
