package unittest_test

import (
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/lrills/helm-unittest/pkg/unittest"
	"github.com/lrills/helm-unittest/pkg/unittest/results"
	"github.com/lrills/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	v2util "k8s.io/helm/pkg/chartutil"
)

// Most used test files
const testSuiteTests string = "_suite_tests"

const testValuesFiles = "../../test/data/services_values.yaml"
const testTestFiles string = "tests/*_test.yaml"
const testTestFailedFiles string = "tests_failed/*_test.yaml"

const testV2InvalidBasicChart string = "../../test/data/v2/invalidbasic"
const testV2BasicChart string = "../../test/data/v2/basic"
const testV2WithSubChart string = "../../test/data/v2/with-subchart"
const testV2WithSubFolderChart string = "../../test/data/v2/with-subfolder"
const testV3InvalidBasicChart string = "../../test/data/v3/invalidbasic"
const testV3BasicChart string = "../../test/data/v3/basic"
const testV3WithSubChart string = "../../test/data/v3/with-subchart"
const testV3WithSubFolderChart string = "../../test/data/v3/with-subfolder"
const testV3WithSubSubFolderChart string = "../../test/data/v3/with-subsubcharts"
const testV3WithFailingTemplateChart string = "../../test/data/v3/failing-template"
const testV3WithSchemaChart string = "../../test/data/v3/with-schema"

var tmpdir, _ = ioutil.TempDir("", testSuiteTests)

func makeTestSuiteResultSnapshotable(result *results.TestSuiteResult) *results.TestSuiteResult {

	for _, test := range result.TestsResult {
		test.Duration, _ = time.ParseDuration("0s")
	}

	return result
}

func validateTestResultAndSnapshots(
	t *testing.T,
	suiteResult *results.TestSuiteResult,
	succeed bool,
	displayName string,
	testResultCount int,
	snapshotCreateCount, snapshotTotalCount, snapshotFailedCount, snapshotVanishedCount uint) {

	a := assert.New(t)
	cupaloy.SnapshotT(t, makeTestSuiteResultSnapshotable(suiteResult))

	a.Equal(succeed, suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(testResultCount, len(suiteResult.TestsResult))
	a.Equal(displayName, suiteResult.DisplayName)

	a.Equal(snapshotCreateCount, suiteResult.SnapshotCounting.Created)
	a.Equal(snapshotTotalCount, suiteResult.SnapshotCounting.Total)
	a.Equal(snapshotFailedCount, suiteResult.SnapshotCounting.Failed)
	a.Equal(snapshotVanishedCount, suiteResult.SnapshotCounting.Vanished)
}

func TestV2ParseTestSuiteUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v2/invalidbasic/tests/deployment_test.yaml", "basic", false, []string{})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV2ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v2/basic/tests/deployment_test.yaml", "basic", true, []string{})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"configmap.yaml", "deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV2ParseTestSuiteFileWithOverrideValuesOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v2/basic/tests/deployment_test.yaml", "basic", true, []string{testValuesFiles})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"configmap.yaml", "deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	a.Equal(2, len(suite.Values)) // Expect images and additional_values.yaml
}

func TestV2ParseTestSuiteFileInSubfolderOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v2/with-subfolder/tests/service_test.yaml", "with-subfolder", true, []string{})

	a.Nil(err)
	a.Equal("test service", suite.Name)
	a.Equal([]string{"webserver/service.yaml"}, suite.Templates)
	a.Equal("should pass", suite.Tests[0].Name)
	a.Equal("should render right if values given", suite.Tests[1].Name)
}

func TestV2RunSuiteWithMultipleTemplatesWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	suiteDoc := `
suite: validate metadata
templates:
  - configmap.yaml
  - deployment.yaml
  - ingress.yaml
  - service.yaml
tests:
  - it: should pass all metadata
    set:
      ingress.enabled: true
    asserts:
      - matchRegex:
          path: metadata.name
          pattern: ^RELEASE-NAME-basic
      - equal:
          path: metadata.labels.app
          value: basic
      - matchRegex:
          path: metadata.labels.chart
          pattern: ^basic-
      - equal:
          path: metadata.labels.release
          value: RELEASE-NAME
      - equal:
          path: metadata.labels.heritage
          value: Tiller
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_multple_template_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "validate metadata", 1, 5, 5, 0, 0)
}

func TestV2RunSuiteWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	suiteDoc := `
suite: test suite name
templates:
  - configmap.yaml
  - deployment.yaml
tests:
  - it: should pass
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_suite_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 2, 2, 0, 0)
}

func TestV2RunSuiteWithOverridesWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	suiteDoc := `
suite: test suite name
templates:
  - crd_backup.yaml
release:
  name: my-release
  namespace: my-namespace
  revision: 1
  upgrade: true
capabilities:
  majorVersion: 1
  minorVersion: 12
  apiVersions:
    - br.dev.local/v1
tests:
  - it: should pass
    asserts:
      - hasDocuments:
          count: 1
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_suite_override_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 1, 1, 0, 0)
}

func TestV2RunSuiteWhenFail(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	suiteDoc := `
suite: test suite name
templates:
  - configmap.yaml
  - deployment.yaml
tests:
  - it: should fail
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Pod
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, false, "test suite name", 1, 0, 0, 0, 0)
}

func TestV2RunSuiteNameOverrideFail(t *testing.T) {
	c, _ := v2util.Load(testV2BasicChart)
	suiteDoc := `
suite: test suite name too long
templates:
  - configmap.yaml
  - deployment.yaml
tests:
  - it: should fail as nameOverride is too long
    set:
      nameOverride: too-long-of-a-name-override-that-should-fail-the-template-immediately
    template: deployment.yaml
    asserts:
      - failedTemplate:
          errorMessage: nameOverride cannot be longer than 20 characters
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_nameoverride_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name too long", 1, 0, 0, 0, 0)
}

func TestV2RunSuiteWithSubfolderWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2WithSubFolderChart)
	suiteDoc := `
suite: test suite name
templates:
  - db/deployment.yaml
  - webserver/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_subfolder_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 2, 2, 0, 0)
}

func TestV2RunSuiteWithSubChartsWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2WithSubChart)
	suiteDoc := `
suite: test suite with subchart
templates:
  - charts/postgresql/templates/deployment.yaml 
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_subchart_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 1, 1, 1, 0, 0)
}

func TestV2RunSuiteWithSubChartsWithAliasWhenPass(t *testing.T) {
	c, _ := v2util.Load(testV2WithSubChart)
	suiteDoc := `
suite: test suite with subchart
templates:
  - charts/postgresql/templates/pvc.yaml 
  - charts/another-postgresql/templates/pvc.yaml
tests:
  - it: should both pass
    asserts:
      - equal:
          path: kind
          value: PersistentVolumeClaim
      - matchSnapshot: {}
  - it: should no pvc for alias
    set:
      another-postgresql.persistence.enabled: false
    template: charts/another-postgresql/templates/pvc.yaml
    asserts:
      - hasDocuments:
          count: 0
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v2_subchartwithalias_test.yaml"), false)
	suiteResult := testSuite.RunV2(c, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 2, 2, 2, 0, 0)
}

func TestV3ParseTestSuiteUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_test.yaml", "basic", false, []string{})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV3ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{})

	a.Nil(err)
	a.Equal(suite.Name, "test deployment")
	a.Equal(suite.Templates, []string{"configmap.yaml", "deployment.yaml"})
	a.Equal(suite.Tests[0].Name, "should pass all kinds of assertion")
}

func TestV3ParseTestSuiteFileWithOverrideValuesOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{testValuesFiles})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"configmap.yaml", "deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	a.Equal(2, len(suite.Values)) // Expect images and additional_values.yaml
}

func TestV3RunSuiteWithMultipleTemplatesWhenPass(t *testing.T) {
	suiteDoc := `
suite: validate metadata
templates:
  - configmap.yaml
  - deployment.yaml
  - ingress.yaml
  - service.yaml
tests:
  - it: should pass all metadata
    set:
      ingress.enabled: true
    asserts:
      - matchRegex:
          path: metadata.name
          pattern: ^RELEASE-NAME-basic
      - equal:
          path: metadata.labels.app
          value: basic
      - matchRegex:
          path: metadata.labels.chart
          pattern: ^basic-
      - equal:
          path: metadata.labels.release
          value: RELEASE-NAME
      - equal:
          path: metadata.labels.heritage
          value: Helm
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_multiple_template_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "validate metadata", 1, 5, 5, 0, 0)
}

func TestV3RunSuiteWhenPass(t *testing.T) {
	suiteDoc := `
suite: test suite name
templates:
  - configmap.yaml
  - deployment.yaml
tests:
  - it: should pass
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 2, 2, 0, 0)
}

func TestV3RunSuiteWithOverridesWhenPass(t *testing.T) {
	suiteDoc := `
suite: test suite name
templates:
  - crd_backup.yaml
release:
  name: my-release
  namespace: my-namespace
  revision: 1
  upgrade: true
capabilities:
  majorVersion: 1
  minorVersion: 10
  apiVersions:
    - br.dev.local/v2
tests:
  - it: should pass
    capabilities:
      majorVersion: 1
      minorVersion: 12
      apiVersions:
        - br.dev.local/v1
    asserts:
      - hasDocuments:
          count: 1
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_suite_override_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 1, 1, 0, 0)
}

func TestV3RunSuiteWhenFail(t *testing.T) {
	suiteDoc := `
suite: test suite name
templates:
  - configmap.yaml
  - deployment.yaml
tests:
  - it: should fail
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Pod
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, false, "test suite name", 1, 0, 0, 0, 0)
}

func TestV3RunSuiteWithSubfolderWhenPass(t *testing.T) {
	suiteDoc := `
suite: test suite name
templates:
  - db/deployment.yaml
  - webserver/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subfolder_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubFolderChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name", 1, 2, 2, 0, 0)
}

func TestV3RunSuiteWithSubChartsWhenPass(t *testing.T) {
	suiteDoc := `
suite: test suite with subchart
templates:
  - charts/postgresql/templates/deployment.yaml
tests:
  - it: should pass
    asserts:
      - equal:
          path: kind
          value: Deployment
      - matchSnapshot: {}
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subchart_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 1, 1, 1, 0, 0)
}

func TestV3RunSuiteWithSubChartsWithAliasWhenPass(t *testing.T) {
	suiteDoc := `
suite: test suite with subchart
templates:
  - charts/postgresql/templates/pvc.yaml 
  - charts/another-postgresql/templates/pvc.yaml
tests:
  - it: should both pass
    asserts:
      - equal:
          path: kind
          value: PersistentVolumeClaim
      - matchSnapshot: {}
  - it: should no pvc for alias
    set:
      another-postgresql.persistence.enabled: false
    template: charts/another-postgresql/templates/pvc.yaml
    asserts:
      - hasDocuments:
          count: 0
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subchartwithalias_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 2, 2, 2, 0, 0)
}

func TestV3RunSuiteNameOverrideFail(t *testing.T) {
	suiteDoc := `
suite: test suite name too long
templates:
  - deployment.yaml
tests:
  - it: should fail as nameOverride is too long
    set:
      nameOverride: too-long-of-a-name-override-that-should-fail-the-template-immediately
    asserts:
      - failedTemplate:
          errorMessage: nameOverride cannot be longer than 20 characters
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_nameoverride_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name too long", 1, 0, 0, 0, 0)
}
