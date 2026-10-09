package unittest_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
)

func newMap(apiVersion, kind, namespace, name string) map[string]any {
	return map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]any{
			"namespace": namespace,
			"name":      name,
		},
	}
}

func TestKubernetesFakeClientProvider(t *testing.T) {
	k := KubernetesFakeClientProvider{
		Scheme:  map[string]KubernetesFakeKindProps{"v1/Pod": {ShouldErr: nil, Gvr: schema.GroupVersionResource{Resource: "pods", Version: "v1"}, Namespaced: true}},
		Objects: []map[string]any{newMap("v1", "Pod", "default", "unittest")},
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

func secretClientFor(t *testing.T, secret map[string]any) *unstructured.Unstructured {
	t.Helper()

	k := KubernetesFakeClientProvider{
		Scheme:  map[string]KubernetesFakeKindProps{"v1/Secret": {Gvr: schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, Namespaced: true}},
		Objects: []map[string]any{secret},
	}

	client, _, err := k.GetClientFor("v1", "Secret")
	assert.NoError(t, err)

	item, err := client.Namespace("default").Get(context.Background(), "unittest", v1.GetOptions{})
	assert.NoError(t, err)

	return item
}

func TestKubernetesFakeClientProviderSecretStringDataIsEncoded(t *testing.T) {
	secret := newMap("v1", "Secret", "default", "unittest")
	secret["stringData"] = map[string]any{"whatever": "whatever value", "port": 8080}

	item := secretClientFor(t, secret)

	assert.Equal(t, map[string]any{
		"whatever": "d2hhdGV2ZXIgdmFsdWU=",
		"port":     "ODA4MA==",
	}, item.Object["data"])
	assert.NotContains(t, item.Object, "stringData")

	// the original test definition must not be mutated
	assert.Contains(t, secret, "stringData")
	assert.NotContains(t, secret, "data")
}

func TestKubernetesFakeClientProviderSecretStringDataOverridesData(t *testing.T) {
	secret := newMap("v1", "Secret", "default", "unittest")
	secret["data"] = map[string]any{"kept": "a2VwdA==", "overridden": "b2xk"}
	secret["stringData"] = map[string]any{"overridden": "new"}

	item := secretClientFor(t, secret)

	assert.Equal(t, map[string]any{
		"kept":       "a2VwdA==",
		"overridden": "bmV3",
	}, item.Object["data"])
}

func TestKubernetesFakeClientProviderSecretWithEmptyStringDataDropsTheField(t *testing.T) {
	secret := newMap("v1", "Secret", "default", "unittest")
	secret["data"] = map[string]any{"kept": "a2VwdA=="}
	secret["stringData"] = nil

	item := secretClientFor(t, secret)

	assert.Equal(t, map[string]any{"kept": "a2VwdA=="}, item.Object["data"])
	assert.NotContains(t, item.Object, "stringData")
}

func TestKubernetesFakeClientProviderSecretWithUnexpectedStringDataIsUntouched(t *testing.T) {
	secret := newMap("v1", "Secret", "default", "unittest")
	secret["data"] = map[string]any{"kept": "a2VwdA=="}
	secret["stringData"] = "not a map"

	item := secretClientFor(t, secret)

	assert.Equal(t, map[string]any{"kept": "a2VwdA=="}, item.Object["data"])
	assert.Equal(t, "not a map", item.Object["stringData"])
}

func TestKubernetesFakeClientProviderSecretWithoutStringDataIsUntouched(t *testing.T) {
	secret := newMap("v1", "Secret", "default", "unittest")
	secret["data"] = map[string]any{"whatever": "d2hhdGV2ZXIgdmFsdWU="}

	item := secretClientFor(t, secret)

	assert.Equal(t, map[string]any{"whatever": "d2hhdGV2ZXIgdmFsdWU="}, item.Object["data"])
}

func TestKubernetesFakeClientProviderStringDataOnlyConvertedForCoreSecret(t *testing.T) {
	stringData := map[string]any{"whatever": "whatever value"}

	custom := newMap("example.com/v1", "Secret", "default", "unittest")
	custom["stringData"] = stringData

	k := KubernetesFakeClientProvider{
		Scheme:  map[string]KubernetesFakeKindProps{"example.com/v1/Secret": {Gvr: schema.GroupVersionResource{Group: "example.com", Resource: "secrets", Version: "v1"}, Namespaced: true}},
		Objects: []map[string]any{custom},
	}

	client, _, err := k.GetClientFor("example.com/v1", "Secret")
	assert.NoError(t, err)

	item, err := client.Namespace("default").Get(context.Background(), "unittest", v1.GetOptions{})
	if assert.NoError(t, err) {
		assert.Equal(t, stringData, item.Object["stringData"])
		assert.NotContains(t, item.Object, "data")
	}
}
