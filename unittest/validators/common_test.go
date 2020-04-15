package validators_test

import (
	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/snapshot"
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

// Most used test files
const testOutputFile string = "../__fixtures__/output/test_output.xml"
const testV2BasicChart string = "../__fixtures__/v2/basic"
const testV2WithSubChart string = "../__fixtures__/v2/with-subchart"
const testV2WithSubFolderChart string = "../__fixtures__/v2/with-subfolder"
const testV3BasicChart string = "../__fixtures__/v3/basic"
const testV3WithSubChart string = "../__fixtures__/v3/with-subchart"
const testV3WithSubFolderChart string = "../__fixtures__/v3/with-subfolder"
