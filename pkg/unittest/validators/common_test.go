package validators_test

import (
	"github.com/lrills/helm-unittest/internal/common"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/mock"
	yaml "gopkg.in/yaml.v2"
)

func makeManifest(doc string) common.K8sManifest {
	manifest := common.K8sManifest{}
	yaml.Unmarshal([]byte(doc), &manifest)
	return manifest
}

type mockSnapshotComparer struct {
	mock.Mock
}

func (m *mockSnapshotComparer) CompareToSnapshot(content interface{}) *snapshot.CompareResult {
	args := m.Called(content)
	return args.Get(0).(*snapshot.CompareResult)
}
