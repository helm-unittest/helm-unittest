package validators_test

import (
	"github.com/helm-unittest/helm-unittest/internal/common"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/mock"
)

func makeManifest(doc string) common.K8sManifest {
	return common.TrustedUnmarshalYAML(doc)
}

type mockSnapshotComparer struct {
	mock.Mock
}

func (m *mockSnapshotComparer) CompareToSnapshot(content any, optFns ...func(options *snapshot.CacheOptions) error) *snapshot.CompareResult {
	args := m.Called(content)
	return args.Get(0).(*snapshot.CompareResult)
}

func (m *mockSnapshotComparer) CompareToSnapshotYAML(contents []any, optFns ...func(options *snapshot.CacheOptions) error) *snapshot.CompareResult {
	args := m.Called(contents)
	return args.Get(0).(*snapshot.CompareResult)
}
