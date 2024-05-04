package unittest

import (
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
	Objects []map[string]interface{}           `yaml:"objects"`
}

func (p *KubernetesFakeClientProvider) GetClientFor(apiVersion, kind string) (dynamic.NamespaceableResourceInterface, bool, error) {
	props := p.Scheme[path.Join(apiVersion, kind)]
	if props.ShouldErr != nil {
		return nil, false, props.ShouldErr
	}

	return fake.NewSimpleDynamicClient(runtime.NewScheme(), convertRuntimeObject(p.Objects)...).Resource(props.Gvr), props.Namespaced, nil
}

func convertRuntimeObject(input []map[string]interface{}) []runtime.Object {
	result := make([]runtime.Object, len(input))

	for k, v := range input {
		result[k] = &unstructured.Unstructured{Object: v}
	}

	return result
}
