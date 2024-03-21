package unittest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newMap(apiVersion, kind, namespace, name string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
}

func TestKubernetesFakeClientProvider(t *testing.T) {
	k := kubernetesFakeClientProvider{
		Scheme:  map[string]kubernetesFakeKindProps{"v1/Pod": {ShouldErr: nil, Gvr: schema.GroupVersionResource{Resource: "pods", Version: "v1"}, Namespaced: true}},
		Objects: []map[string]interface{}{newMap("v1", "Pod", "default", "unittest")},
	}

	client, namespaced, err := k.GetClientFor("v1", "Pod")
	assert.NoError(t, err)
	assert.True(t, namespaced)
	assert.NotNil(t, client)

	item, err := client.Namespace("default").Get(context.Background(), "unittest", v1.GetOptions{})
	if assert.NoError(t, err) {
		assert.Equal(t, item.GetNamespace(), "default")
		assert.Equal(t, item.GetName(), "unittest")
		assert.Equal(t, item.GetAPIVersion(), "v1")
		assert.Equal(t, item.GetKind(), "Pod")
	}

	_, err = client.Namespace("default").Get(context.Background(), "notexisting", v1.GetOptions{})
	assert.Error(t, err)
}
