package unittest_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/helm-unittest/helm-unittest/internal/common"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

// Most used test files
const testSuiteTests string = "_suite_tests"

const testValuesFiles = "../../test/data/services_values.yaml"
const testExternalValuesFiles = "../../test/data/external/ns_values.yaml"
const testExternalSubTestFiles = "../../test/data/external/*.yaml"
const testExternalTestFiles = "../../test/data/external/tests/*.yaml"
const testTestFiles string = "tests/*_test.yaml"
const testTestFailedFiles string = "tests_failed/*_test.yaml"

const testV3InvalidBasicChart string = "../../test/data/v3/invalidbasic"
const testV3BasicChart string = "../../test/data/v3/basic"
const testV3FullSnapshotChart string = "../../test/data/v3/full-snapshot"
const testV3WithSubChart string = "../../test/data/v3/with-subchart"
const testV3WithSubFolderChart string = "../../test/data/v3/with-subfolder"
const testV3WithSubSubFolderChart string = "../../test/data/v3/with-subsubcharts"
const testV3WithFilesChart string = "../../test/data/v3/with-files"
const testV3WithFailingTemplateChart string = "../../test/data/v3/failing-template"
const testV3WithSchemaChart string = "../../test/data/v3/with-schema"
const testV3WithPackagedChart string = "../../test/data/v3/with-packaged-0.1.0.tgz"
const testV3WithPackagedSubChart string = "../../test/data/v3/with-subchart/charts/postgresql-0.8.3.tgz"
const testV3GlobalDoubleChart string = "../../test/data/v3/global-double-setting"
const testV3WithHelmTestsChart string = "../../test/data/v3/with-helm-tests"
const testV3WitSamenameSubSubChart string = "../../test/data/v3/with-samenamesubsubcharts"
const testV3WithDocumentSelectorChart string = "../../test/data/v3/with-document-select"
const testV3WithFakeK8sClientChart string = "../../test/data/v3/with-k8s-fake-client"
const testV3WithPostRendererChart string = "../../test/data/v3/with-post-renderer"

var tmpdir, _ = os.MkdirTemp("", testSuiteTests)

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

// Helper metheod for the render process
func getExpectedRenderedTestSuites(customSnapshotIds bool, t *testing.T) map[string]*TestSuite {
	// multiple_suites_snapshot.yaml assertions
	createSnapshotTestYaml := func(env string) string {
		return fmt.Sprintf(`
it: manifest should match snapshot
set:
    env: %s
asserts:
    - matchSnapshot: {}`, env)
	}
	snapshotDevTest := TestJob{}
	common.YmlUnmarshalTestHelper(createSnapshotTestYaml("dev"), &snapshotDevTest, t)
	snapshotProdTest := TestJob{}
	common.YmlUnmarshalTestHelper(createSnapshotTestYaml("prod"), &snapshotProdTest, t)
	// multiple_test_suites.yaml assertions
	crateMultipleTestSuitesYaml := func(env string) string {
		return fmt.Sprintf(`
it: validate base64 encoded value
set:
    postgresql:
      postgresPassword: %s
    another-postgresql:
      postgresPassword: password
asserts:
    - isKind:
        of: Secret
    - hasDocuments:
        count: 1
    - equal:
        path: data.postgres-password
        value: %s
        decodeBase64: true`, env, env)
	}
	multipleTestSuitesDevTest := TestJob{}
	common.YmlUnmarshalTestHelper(crateMultipleTestSuitesYaml("dev"), &multipleTestSuitesDevTest, t)
	multipleTestSuitesProdTest := TestJob{}
	common.YmlUnmarshalTestHelper(crateMultipleTestSuitesYaml("prod"), &multipleTestSuitesProdTest, t)
	// multiple_tests_test.yaml assertions
	var secretNameEqualsYaml = func(env string) string {
		return fmt.Sprintf(`
it: should set tls in for %s
set:
    ingress.enabled: true
    ingress.tls:
      - secretName: %s-my-tls-secret
asserts:
    - equal:
        path: spec.tls
        value:
          - secretName: %s-my-tls-secret`, env, env, env)
	}
	multipleTestsDevTest := TestJob{}
	common.YmlUnmarshalTestHelper(secretNameEqualsYaml("dev"), &multipleTestsDevTest, t)
	multipleTestsProdTest := TestJob{}
	common.YmlUnmarshalTestHelper(secretNameEqualsYaml("prod"), &multipleTestsProdTest, t)
	const multipleTestsFirstTestYaml = `
it: should render nothing if not enabled
asserts:
    - hasDocuments:
        count: 0`
	multipleTestsFirstTest := TestJob{}
	common.YmlUnmarshalTestHelper(multipleTestsFirstTestYaml, &multipleTestsFirstTest, t)

	// Set up snapshotId values
	// Note, this is completely based on the order of the yaml in a single suite template file
	var (
		multipleTestSuiteDevSnapshotId       string
		multipleTestSuiteProdSnapshotId      string
		multipleSuiteSnapshotsDevSnapshotId  string
		multipleSuiteSnapshotsProdSnapshotId string
		multipleTestsSnapshotId              string
	)
	if customSnapshotIds {
		multipleTestSuiteDevSnapshotId = "dev"
		multipleTestSuiteProdSnapshotId = "prod"
		multipleSuiteSnapshotsDevSnapshotId = "dev"
		multipleSuiteSnapshotsProdSnapshotId = "prod"
		multipleTestsSnapshotId = "all"
	} else {
		multipleTestSuiteDevSnapshotId = "0"
		multipleTestSuiteProdSnapshotId = "1"
		multipleSuiteSnapshotsDevSnapshotId = "0"
		multipleSuiteSnapshotsProdSnapshotId = "1"
		multipleTestsSnapshotId = "0"
	}

	return map[string]*TestSuite{
		"multiple test suites dev": {
			Templates:  []string{"charts/postgresql/templates/secrets.yaml"},
			SnapshotId: multipleTestSuiteDevSnapshotId,
			Tests: []*TestJob{
				&multipleTestSuitesDevTest,
			},
		},
		"multiple test suites prod": {
			Templates:  []string{"charts/postgresql/templates/secrets.yaml"},
			SnapshotId: multipleTestSuiteProdSnapshotId,
			Tests: []*TestJob{
				&multipleTestSuitesProdTest,
			},
		},
		"multiple test suites snapshot dev": {
			Templates:  []string{"templates/service.yaml"},
			SnapshotId: multipleSuiteSnapshotsDevSnapshotId,
			Tests: []*TestJob{
				&snapshotDevTest,
			},
		},
		"multiple test suites snapshot prod": {
			Templates:  []string{"templates/service.yaml"},
			SnapshotId: multipleSuiteSnapshotsProdSnapshotId,
			Tests: []*TestJob{
				&snapshotProdTest,
			},
		},
		"multiple tests": {
			Templates:  []string{"templates/ingress.yaml"},
			SnapshotId: multipleTestsSnapshotId,
			Tests: []*TestJob{
				&multipleTestsFirstTest,
				&multipleTestsDevTest,
				&multipleTestsProdTest,
			},
		},
	}
}

func TestV3ParseTestSuite_FileNotExist(t *testing.T) {
	_, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment.yaml", "basic", false, []string{})
	assert.Error(t, err)
}

func TestV3ParseTestSuiteUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_test.yaml", "basic", false, []string{})

	a.Nil(err)
	a.Len(suites, 2)
	for _, suite := range suites {
		a.Equal("test deployment", suite.Name)
		a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
		a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	}
}

func TestV3ParseTestSuiteUnstrictNoTestsFileFail(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_notests_test.yaml", "basic", false, []string{})

	a.NotNil(err)
	a.EqualError(err, "no tests found")
	a.Len(suites, 1)
	for _, suite := range suites {
		a.Equal("test deployment", suite.Name)
		a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
	}
}

func TestV3ParseTestSuiteUnstrictNoAssertsFileFail(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_noasserts_test.yaml", "basic", false, []string{})

	a.NotNil(err)
	a.EqualError(err, "no asserts found")
	a.Len(suites, 1)
	for _, suite := range suites {
		a.Equal("test deployment", suite.Name)
		a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
		a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	}
}

func TestV3ParseTestSuiteStrictFileError(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_test.yaml", "basic", true, []string{})

	a.NotNil(err)
	a.EqualError(err, "yaml: unmarshal errors:\n  line 6: field documents not found in type unittest.TestJob")
	a.Len(suites, 2)
	for _, suite := range suites {
		a.Equal("test deployment", suite.Name)
		a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
		a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	}
}

func TestV3ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{})

	a.Nil(err)
	for _, suite := range suites {
		a.Equal(suite.Name, "test deployment")
		a.Equal(suite.Templates, []string{"templates/configmap.yaml", "templates/deployment.yaml"})
		a.Equal(suite.Tests[0].Name, "should pass all kinds of assertion")
	}
}

func TestV3ParseTestSuiteFileWithOverrideValuesOk(t *testing.T) {
	a := assert.New(t)
	suites, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{testValuesFiles})

	a.Nil(err)
	for _, suite := range suites {
		a.Equal("test deployment", suite.Name)
		a.Equal([]string{"templates/configmap.yaml", "templates/deployment.yaml"}, suite.Templates)
		a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
		a.Equal(1, len(suite.Values)) // Expect services_values.yaml
	}
}

func TestV3RenderSuitesUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", false, []string{}, map[string]interface{}{
		"unexpectedField": false,
	})

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(false, t)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite"+suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name ("+suite.Name+") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name ("+suite.Name+") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name ("+suite.Name+") mismatched tests")
	}
}

func TestV3RenderSuitesStrictFileFail(t *testing.T) {
	a := assert.New(t)
	_, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{}{
		"unexpectedField": true,
	})

	a.NotNil(err)
	a.ErrorContains(err, "field something not found in type unittest.TestSuite")
}

func TestV3RenderSuites_InvalidDirectory(t *testing.T) {
	a := assert.New(t)
	_, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart-not-exist", "basic", true, []string{}, map[string]interface{}{
		"unexpectedField": true,
	})
	a.Error(err)
	a.ErrorIs(err, os.ErrNotExist)
}

func TestV3RenderSuites_LoadError(t *testing.T) {
	a := assert.New(t)
	tmp := t.TempDir()
	chartPath := path.Join(tmp, "basic")
	_ = os.MkdirAll(chartPath, 0755)
	chart := `
name: basic
`
	a.NoError(writeToFile(chart, path.Join(chartPath, "Chart.yaml")))
	defer os.RemoveAll(chartPath)

	_, err := RenderTestSuiteFiles(chartPath, "basic", false, []string{}, nil)
	a.Error(err)
	a.ErrorContains(err, "validation: chart.metadata.version is required")
}

func TestV3RenderSuites_RenderError(t *testing.T) {
	a := assert.New(t)
	tmp := t.TempDir()
	chartPath := path.Join(tmp, "basic")
	_ = os.MkdirAll(chartPath, 0755)
	chart := `
name: basic
version: 1.0.0
`
	deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .BreakV3engine.Render }}-basic
spec:
  replicas: 3
`

	a.NoError(writeToFile(chart, path.Join(chartPath, "Chart.yaml")))
	a.NoError(writeToFile(deployment, path.Join(chartPath, "templates/deployment.yaml")))
	defer os.RemoveAll(chartPath)
	_, err := RenderTestSuiteFiles(chartPath, "basic", false, []string{}, nil)

	a.Error(err)
	a.ErrorContains(err, "executing \"basic/templates/deployment.yaml\" at <.BreakV3engine.Render>")
}

func TestV3RenderSuites_RenderValuesWithIterateAllKeysError(t *testing.T) {
	a := assert.New(t)
	tmp := t.TempDir()
	chartPath := path.Join(tmp, "basic")
	_ = os.MkdirAll(chartPath, 0755)
	chart := `
name: basic
version: 1.0.0
`
	empty_manifest := ``

	a.NoError(writeToFile(chart, path.Join(chartPath, "Chart.yaml")))
	a.NoError(writeToFile(empty_manifest, path.Join(chartPath, "templates/deployment.yaml")))
	defer os.RemoveAll(chartPath)
	values := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	_, err := RenderTestSuiteFiles(chartPath, "basic", false, []string{}, values)
	a.Error(err)
	a.ErrorContains(err, "file did not render a manifest")
}

func TestV3RenderSuitesFailNoSuiteName(t *testing.T) {
	a := assert.New(t)
	_, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{}{
		"includeSuite": false,
	})
	a.NotNil(err)
	a.ErrorContains(err, "helm chart based test suites must include `suite` field")
}

func TestV3RenderSuitesStrictFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, nil)

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(false, t)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite"+suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name ("+suite.Name+") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name ("+suite.Name+") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name ("+suite.Name+") mismatched tests")
	}
}

func TestV3RenderSuitesCustomSnapshotIdOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{}{
		"customSnapshotIds": true,
	})

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(true, t)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite"+suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name ("+suite.Name+") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name ("+suite.Name+") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name ("+suite.Name+") mismatched tests")
	}
}

func TestV3RunSuiteWithNoAssertsShouldFail(t *testing.T) {
	suiteDoc := `
suite: validate empty asserts
tests:
  - it: should fail with no asserts
    asserts:
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_noasserts_template_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, false, "validate empty asserts", 1, 0, 0, 0, 0)
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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_multiple_template_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_suite_override_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subfolder_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubFolderChart, cache, true, "", &results.TestSuiteResult{})

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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subchart_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubChart, cache, true, "", &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 1, 1, 1, 0, 0)
}

func TestV3RunSuiteWithSubChartAliasAndVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with subchart and version override
chart:
  version: 1.2.3
tests:
  - it: should render subchart and alias subchart templates
    templates:
     - charts/another-postgresql/templates/deployment.yaml
     - charts/postgresql/templates/deployment.yaml
    asserts:
     - matchRegex:
        path: metadata.labels["chart"]
        pattern: "(.*-)?postgresql-1.2.3"
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	suiteResult := testSuite.RunV3(testV3WithSubChart, &snapshot.Cache{}, true, "", &results.TestSuiteResult{})
	assert.True(t, suiteResult.Passed)
}

func TestV3RunSuiteWithSubChartsTrimmingWhenPass(t *testing.T) {
	suiteDoc := `
suite: test cert-manager rbac with trimming
templates:
  - charts/cert-manager/templates/rbac.yaml
tests:
  - it: templates
    release:
      name: cert-manager
      namespace: cert-manager
    asserts:
      - notFailedTemplate: {}
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subchartwithtrimming_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubChart, cache, true, "", &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test cert-manager rbac with trimming", 1, 0, 0, 0, 0)
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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_subchartwithalias_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3WithSubChart, cache, true, "", &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite with subchart", 2, 2, 2, 0, 0)
}

func TestV3RunSuiteWithSubChartsWithAliasWithoutChartVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite without subchart version override
templates:
  - charts/postgresql/templates/pvc.yaml
tests:
  - it: should no pvc for alias
    set:
      postgresql.persistence.enabled: true
    asserts:
      - hasDocuments:
          count: 1
      - matchSnapshot: {}
      - equal:
          path: metadata.labels.chart
          value: postgresql-0.8.3
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	suiteResult := testSuite.RunV3(testV3WithSubChart, &snapshot.Cache{}, true, "", &results.TestSuiteResult{})

	assert.Empty(t, testSuite.Chart.AppVersion)
	assert.Empty(t, testSuite.Chart.Version)
	assert.True(t, suiteResult.Passed)
}

func TestV3RunSuiteWithSubChartsWithAliasWithSuiteChartVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with suite version override
templates:
  - charts/postgresql/templates/pvc.yaml
chart:
  version: 0.6.3
tests:
  - it: should no pvc for alias
    set:
      postgresql.persistence.enabled: true
    asserts:
      - hasDocuments:
          count: 1
      - equal:
          path: metadata.labels.chart
          value: postgresql-0.6.3
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	suiteResult := testSuite.RunV3(testV3WithSubChart, &snapshot.Cache{}, true, "", &results.TestSuiteResult{})

	assert.Empty(t, testSuite.Chart.AppVersion)
	assert.Equal(t, testSuite.Chart.Version, "0.6.3")
	assert.True(t, suiteResult.Passed)
}

func TestV3RunSuiteWithSubChartsWithAliasWithJobChartVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with suite version override
templates:
  - charts/postgresql/templates/pvc.yaml
chart:
  version: 0.6.2
tests:
  - it: should no pvc for alias
    set:
      postgresql.persistence.enabled: true
    chart:
        version: 0.7.1
    asserts:
      - hasDocuments:
          count: 1
      - equal:
          path: metadata.labels.chart
          value: postgresql-0.7.1
`
	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	suiteResult := testSuite.RunV3(testV3WithSubChart, &snapshot.Cache{}, true, "", &results.TestSuiteResult{})

	assert.Empty(t, testSuite.Chart.AppVersion)
	assert.Equal(t, testSuite.Chart.Version, "0.6.2")
	assert.True(t, suiteResult.Passed)
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
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_nameoverride_failed_suite_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, "", &results.TestSuiteResult{})

	validateTestResultAndSnapshots(t, suiteResult, true, "test suite name too long", 1, 0, 0, 0, 0)
}

func TestV3ParseTestMultipleSuitesWithSingleSeparator(t *testing.T) {
	suiteDoc := `
suite: first suite without leading triple dashes
templates:
  - deployment.yaml
tests:
  - it: should fail as nameOverride is too long
    set:
      nameOverride: too-long-of-a-name-override-that-should-fail-the-template-immediately
    asserts:
      - failedTemplate:
          errorMessage: nameOverride cannot be longer than 20 characters
---
suite: second suite in same separated with triple dashes
templates:
  - deployment.yaml
tests:
  - it: should fail due to paradox
    set:
      name: first-deployment
    asserts:
      - failedTemplate: {}
`
	a := assert.New(t)
	file := path.Join("_scratch", "multiple-suites-withsingle-separator.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.Remove(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 2)
}

func TestV3ParseTestMultipleSuitesWithSeparatorsAndSetMultilineValue(t *testing.T) {
	suiteDoc := `
---
suite: first test suite for deployment
templates:
  - deployment.yaml
tests:
  - it: should render deployment
    set:
      name: first-deployment
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
---
suite: second suite in same file
templates:
  - deployment.yaml
tests:
  - it: should render second deployment in second suite
    set:
      signing.privateKey: |-
        -----BEGIN PGP PRIVATE KEY BLOCK-----
        {placeholder}
        -----END PGP PRIVATE KEY BLOCK-----
    asserts:
      - containsDocument:
          kind: Deployment
          apiVersion: v1
---
suite: third suite in same file
templates:
  - secret.yaml
tests:
  - it: should render second deployment in second suite
    set:
      signing.privateKey: |-
        -----BEGIN PGP PRIVATE KEY BLOCK-----
        {placeholder}
        -----END PGP PRIVATE KEY BLOCK-----
    asserts:
      - containsDocument:
          kind: Secret
          apiVersion: v1
`
	a := assert.New(t)
	file := path.Join("_scratch", "multiple-suites-with-multiline-value.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 3)
}

func TestV3ParseTestSingleSuitesWithSuiteChartMetadataOverride(t *testing.T) {
	suiteDoc := `
---
suite: test suite with explicit version and appVersion
templates:
  - deployment.yaml
chart:
  appVersion: v1
  version: 1.0.0
tests:
  - it: should render deployment
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-chart-metadata.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "override", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)

	for _, suite := range suites {
		a.Equal("1.0.0", suite.Chart.Version)
		a.Equal("v1", suite.Chart.AppVersion)
	}
}

func TestV3ParseTestSingleSuiteWithTestChartMetadataOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with explicit version and appVersion
templates:
  - deployment.yaml
chart:
  appVersion: v1
  version: 1.0.0
tests:
  - it: should override chart.version
    chart:
      version: 1.0.1
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-chart-metadata.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)

	for _, suite := range suites {
		a.Equal("1.0.0", suite.Chart.Version)
		a.Equal("v1", suite.Chart.AppVersion)
		a.Len(suite.Tests, 1)
		a.Equal("1.0.1", suite.Tests[0].Chart.Version)
		a.Equal("v1", suite.Tests[0].Chart.AppVersion)
	}
}

func TestV3ParseTestSingleSuitesWithMutlipleTestChartMetadataOverride(t *testing.T) {
	suiteDoc := `
suite: test suite without chart metadata
templates:
  - deployment.yaml
tests:
  - it: should override chart metadata
    chart:
      version: 1.0.1
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
  - it: should not override chart metadata
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-chart-metadata.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)

	for _, suite := range suites {
		a.Equal("", suite.Chart.Version)
		a.Equal("", suite.Chart.AppVersion)
		a.Len(suite.Tests, 2)
		a.Equal("1.0.1", suite.Tests[0].Chart.Version)
		a.Equal("", suite.Tests[0].Chart.AppVersion)
		a.Equal("", suite.Tests[1].Chart.Version)
		a.Equal("", suite.Tests[1].Chart.AppVersion)
	}
}

func TestV3ParseTestSingleSuitesWithChartMetadataAndEmptyVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
chart:
  appVersion: v3
tests:
  - it: should not override with empty appVersion
    chart:
      appVersion:
    asserts:
      - equal:
          path: metadata.labels.chart
          value: deployment-test
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-chart-metadata.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)

	for _, suite := range suites {
		a.Equal("v3", suite.Chart.AppVersion)
		a.Len(suite.Tests, 1)
		a.Equal("v3", suite.Tests[0].Chart.AppVersion)
	}
}

func TestV3ParseTestSingleSuitesWithKubeCapabilitiesUnset(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
capabilities:
  apiVersions:
    - autoscaling/v2
tests:
  - it: should not override with empty appVersion
    capabilities:
      apiVersions:
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "unset-test-apiversions.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)
	a.Equal([]string{"autoscaling/v2"}, suites[0].Capabilities.APIVersions)
	a.Equal([]string(nil), suites[0].Tests[0].Capabilities.APIVersions)
}

func TestV3ParseTestSingleSuitesWithKubeCapabilitiesOverrided(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
capabilities:
  apiVersions:
   - autoscaling/v2
tests:
  - it: should not override with empty appVersion
    capabilities:
      apiVersions:
       - autoscaling/v1
       - monitoring.coreos.com/v1
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-apiversions.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)
	a.Equal([]string{"autoscaling/v2"}, suites[0].Capabilities.APIVersions)
	a.Equal([]string{"autoscaling/v1", "monitoring.coreos.com/v1", "autoscaling/v2"}, suites[0].Tests[0].Capabilities.APIVersions)
}

func TestV3ParseTestSingleSuitesShouldNotUnsetSuiteK8sVersions(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
capabilities:
  majorVersion: 1
  minorVersion: 15
tests:
  - it: should not override with empty appVersion
    capabilities:
      majorVersion:
      minorVersion:
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-apiversions.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)
	a.Equal(suites[0].Capabilities.MajorVersion, suites[0].Tests[0].Capabilities.MajorVersion)
	a.Equal(suites[0].Capabilities.MinorVersion, suites[0].Tests[0].Capabilities.MinorVersion)
}

func TestV3ParseTestSingleSuitesWithSuiteK8sVersionOverride(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
capabilities:
  majorVersion: 1
  minorVersion: 15
tests:
  - it: should not override with empty appVersion
    capabilities:
      majorVersion:
      minorVersion: 10
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "override-test-apiversions.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 1)
	a.Equal(suites[0].Capabilities.MajorVersion, suites[0].Tests[0].Capabilities.MajorVersion)
	a.NotEqual(suites[0].Capabilities.MinorVersion, suites[0].Tests[0].Capabilities.MinorVersion)
	a.Equal("15", suites[0].Capabilities.MinorVersion)
	a.Equal("10", suites[0].Tests[0].Capabilities.MinorVersion)
}

func TestV3ParseTestMultipleSuitesWithK8sVersionOverrides(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
capabilities:
  majorVersion: 1
  minorVersion: 15
  apiVersions:
   - v1
tests:
  - it: should keep majorVersion, minorVersion and keep apiVersions
    capabilities:
      majorVersion:
      minorVersion: 10
    asserts:
      - hasDocuments:
          count: 1
---
suite: second suite in same file
templates:
  - deployment.yaml
capabilities:
  majorVersion: 4
  minorVersion: 13
tests:
  - it: should keep majorVersion, unset apiVersion and override minorVersion
    capabilities:
      majorVersion:
      minorVersion: 11
      apiVersions:
    asserts:
      - hasDocuments:
          count: 1
---
suite: third suite in same file
templates:
  - deployment.yaml
capabilities:
  majorVersion: 3
  minorVersion: 11
  apiVersions:
   - v1
tests:
  - it: should override majorVersion, keep minorVersion and extend apiVersions
    capabilities:
      majorVersion: 1
      minorVersion:
      apiVersions:
       - extensions/v1beta1
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "multiple-capabilities-modifications.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Nil(err)
	a.Len(suites, 3)
	// first
	a.Equal(suites[0].Capabilities.MajorVersion, suites[0].Tests[0].Capabilities.MajorVersion)
	a.NotEqual(suites[0].Capabilities.MinorVersion, suites[0].Tests[0].Capabilities.MinorVersion)
	a.Equal(suites[0].Capabilities.APIVersions, suites[0].Tests[0].Capabilities.APIVersions)
	a.Equal("15", suites[0].Capabilities.MinorVersion)
	a.Equal("10", suites[0].Tests[0].Capabilities.MinorVersion)
	// second
	a.Equal(suites[1].Capabilities.MajorVersion, suites[1].Tests[0].Capabilities.MajorVersion)
	a.Equal("11", suites[1].Tests[0].Capabilities.MinorVersion)
	// third
	a.NotEqual(suites[2].Capabilities.MajorVersion, suites[2].Tests[0].Capabilities.MajorVersion)
	a.Equal("1", suites[2].Tests[0].Capabilities.MajorVersion)
	a.NotEqual(len(suites[2].Capabilities.APIVersions), len(suites[2].Tests[0].Capabilities.APIVersions))
}

func TestV3ParseTestMultipleSuitesWithNotSupportedAssert(t *testing.T) {
	suiteDoc := `
suite: test suite with assert that not supported
templates:
  - deployment.yaml
tests:
  - it: should error when not supported assert is found
    asserts:
      - notSupportedAssert:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "assert-not-supported.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	_, err := ParseTestSuiteFile(file, "basic", true, []string{})

	a.Error(err)
	a.ErrorContains(err, "Assertion type `notSupportedAssert` is invalid")
}

func TestV3ParseTestMultipleSuitesDocumentSelectorWithPoisonInAssertIgnored(t *testing.T) {
	suiteDoc := `
suite: test suite with assert that not supported
templates:
  - "*.yaml"
tests:
  - it: should error when not supported assert is found
    documentSelector:
     skipEmptyTemplates: true # this is a poison pill
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "assert-not-supported.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	_, err := ParseTestSuiteFile(file, "basic", true, []string{})
	a.NoError(err)
}

func TestV3ParseTestMultipleSuitesDocumentSelectorWithPoisonInTestNotIgnored(t *testing.T) {
	suiteDoc := `
suite: test suite with assert that not supported
templates:
  - deployment.yaml
tests:
  - it: should error when not supported assert is found
    asserts:
      - hasDocuments:
          count: 1
        documentSelector:
          skipEmptyTemplates: true
`
	a := assert.New(t)
	file := path.Join("_scratch", "assert-not-supported.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	_, err := ParseTestSuiteFile(file, "basic", true, []string{})
	a.Error(err)
	a.ErrorContains(err, "empty 'documentSelector.path' not supported")
}

func TestV3ParseTestMultipleSuites_With_FailFast(t *testing.T) {
	suiteDoc := `
suite: test suite with partial chart metadata
templates:
  - deployment.yaml
tests:
  - it: should execute this test
    asserts:
      - hasDocuments:
          count: 1
---
suite: second suite failed test and fail fast
templates:
  - deployment.yaml
tests:
  - it: should fail and trigger fail fast
    asserts:
      - hasDocuments:
          count: 2
---
suite: third suite in same file with fail fast triggered in previous suite
templates:
  - deployment.yaml
tests:
  - it: should not execute this test
    asserts:
      - hasDocuments:
          count: 1
`
	a := assert.New(t)
	file := path.Join("_scratch", "fail-fast.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})
	a.Nil(err)
	a.Len(suites, 3)

	testSuite := TestSuite{}
	common.YmlUnmarshalTestHelper(suiteDoc, &testSuite, t)

	suiteResult := testSuite.RunV3(testV3BasicChart, &snapshot.Cache{}, true, "", &results.TestSuiteResult{})

	assert.True(t, suiteResult.FailFast)
	assert.False(t, suiteResult.Passed)
}

func TestV3RunSuiteWithSuite_With_EmptyTestJobs(t *testing.T) {
	testSuite := TestSuite{}
	testSuite.Tests = []*TestJob{
		{
			Name: "first test that is empty",
		},
		{
			Name: "second test that is empty",
		},
	}

	cases := []struct {
		failFast bool
	}{
		{
			failFast: true,
		},
		{
			failFast: false,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("fail fast: %v", tt.failFast), func(t *testing.T) {
			suiteResult := testSuite.RunV3(testV3BasicChart, &snapshot.Cache{}, tt.failFast, "", &results.TestSuiteResult{})
			assert.False(t, suiteResult.Passed)
			assert.True(t, len(suiteResult.TestsResult) == 2)
		})
	}
}

func TestV3MultipleSuitesWithSkip(t *testing.T) {
	suiteDoc := `
---
suite: test skip on suite level
templates:
  - deployment.yaml
skip:
  reason: "This suite is not ready yet"
tests:
  - it: should render deployment
    set:
      name: first-deployment
    asserts:
      - exists:
          path: metadata.labels.chart
---
suite: suite with single skipped test
templates:
  - deployment.yaml
tests:
  - it: should render second deployment in second suite
    skip:
      reason: "skip me"
    asserts:
      - isKind:
          of: Deployment
---
suite: suite with two tests and one skipped test
templates:
  - deployment.yaml
tests:
  - it: should skip test
    skip:
      reason: "skip me"
    asserts:
      - isKind:
          of: Deployment
  - it: should not skip test
    asserts:
      - isKind:
          of: Deployment
---
suite: suite without skip
templates:
  - secret.yaml
tests:
  - it: should render second deployment in second suite
    asserts:
      - isKind:
          of: Secret
---
suite: test skip with minimumVersion on suite level
templates:
  - deployment.yaml
skip:
  minimumVersion: "99.0.0"
tests:
  - it: should be skipped due to minimumVersion
    asserts:
      - exists:
          path: metadata.labels.chart
`

	a := assert.New(t)
	file := path.Join("_scratch", "multiple-suites-with-skip.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	assert.NoError(t, err)
	assert.Len(t, suites, 5)

	for _, s := range suites {
		switch s.Name {
		case "test skip on suite level":
			assert.NotEmpty(t, s.Skip.Reason)
		case "suite with single skipped test":
			assert.NotEmpty(t, s.Skip.Reason)
		case "suite with two tests and one skipped test":
			assert.Empty(t, s.Skip.Reason)
			for _, test := range s.Tests {
				switch test.Name {
				case "should skip test":
					assert.NotEmpty(t, test.Skip.Reason)
				default:
					assert.Empty(t, test.Skip.Reason)
				}
			}
		case "test skip with minimumVersion on suite level":
			assert.NotEmpty(t, s.Skip.Reason)
			assert.Contains(t, s.Skip.Reason, "Test suite requires minimum unittest plugin version 99.0.0")
			// Verify that the Skip.Reason is propagated to all tests
			for _, test := range s.Tests {
				assert.Equal(t, s.Skip.Reason, test.Skip.Reason)
			}
		default:
			assert.Empty(t, s.Skip.Reason)
		}
	}
}

func TestSkipReasonPropagation(t *testing.T) {
	testCases := []struct {
		name         string
		suiteContent string
		expectReason string
	}{
		{
			name: "explicit skip reason",
			suiteContent: `
suite: Test Suite with Skip Reason
templates:
  - deployment.yaml
skip:
  reason: "Explicitly skipped"
tests:
  - it: should test something
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Deployment
`,
			expectReason: "Explicitly skipped",
		},
		{
			name: "minimum version too high",
			suiteContent: `
suite: Test Suite with High MinimumVersion
templates:
  - deployment.yaml
skip:
  minimumVersion: 99.0.0
tests:
  - it: should test something
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Deployment
`,
			expectReason: "Test suite requires minimum unittest plugin version 99.0.0",
		},
		{
			name: "both reason and minimum version",
			suiteContent: `
suite: Test Suite with Both Skip Conditions
templates:
  - deployment.yaml
skip:
  reason: "Should skip regardless of minimum version"
  minimumVersion: 0.0.1
tests:
  - it: should test something
    template: deployment.yaml
    asserts:
      - equal:
          path: kind
          value: Deployment
`,
			expectReason: "Should skip regardless of minimum version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file with the test suite content
			file := path.Join("_scratch", fmt.Sprintf("skip-reason-%s.yaml", tc.name))
			err := writeToFile(tc.suiteContent, file)
			assert.Nil(t, err)
			defer os.RemoveAll(file)

			// Parse the test suite file
			suites, err := ParseTestSuiteFile(file, "basic", true, []string{})
			assert.NoError(t, err)
			assert.Len(t, suites, 1)

			// Check if the skip reason propagates correctly
			testSuite := suites[0]

			// Run the suite
			cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, fmt.Sprintf("skip-reason-snapshot-%s.yaml", tc.name)), false)
			suiteResult := testSuite.RunV3(testV3BasicChart, cache, false, "", &results.TestSuiteResult{})

			// Verify skipped status
			assert.True(t, suiteResult.Skipped)
			assert.True(t, suiteResult.Passed) // Skipped suites should pass

			// Check test jobs have the expected skip reason
			for _, testJob := range testSuite.Tests {
				assert.Contains(t, testJob.Skip.Reason, tc.expectReason)
			}

			// Verify all test results are marked as skipped
			for _, testResult := range suiteResult.TestsResult {
				assert.True(t, testResult.Skipped)
			}
		})
	}
}

func TestV3RunSuiteWithSkipTests(t *testing.T) {
	testSuite := TestSuite{}
	testSuite.Tests = []*TestJob{
		{
			Name: "should skip",
			Skip: struct {
				Reason string `yaml:"reason"`
			}{Reason: "skip me"},
		},
		{
			Name: "should skip test",
		},
	}

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "non-empty-snapshot.yaml"), false)
	cases := []struct {
		failFast bool
	}{
		{
			failFast: true,
		},
		{
			failFast: false,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("fail fast: %v", tt.failFast), func(t *testing.T) {
			suiteResult := testSuite.RunV3(testV3BasicChart, cache, tt.failFast, "", &results.TestSuiteResult{})

			assert.False(t, suiteResult.Skipped)
			assert.False(t, suiteResult.Passed)
		})
	}
}

func TestV3RunSuiteWithSuiteLevelSkip(t *testing.T) {
	testSuite := TestSuite{
		Skip: struct {
			Reason         string `yaml:"reason"`
			MinimumVersion string `yaml:"minimumVersion"`
		}{Reason: "skip suite", MinimumVersion: ""},
	}
	testSuite.Tests = []*TestJob{
		{
			Name: "first. should be skipped",
		},
		{
			Name: "second. should be skipped",
		},
	}

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "non-empty-snapshot.yaml"), false)

	cases := []struct {
		failFast bool
	}{
		{
			failFast: true,
		},
		{
			failFast: false,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("fail fast: %v", tt.failFast), func(t *testing.T) {
			suiteResult := testSuite.RunV3(testV3BasicChart, cache, tt.failFast, "", &results.TestSuiteResult{})

			assert.True(t, suiteResult.Skipped)
			assert.True(t, suiteResult.Passed)
		})
	}
}

func TestVersionMeetsMinimum(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		minimumVersion string
		expected       bool
	}{
		{
			name:           "equal versions",
			currentVersion: "0.8.0",
			minimumVersion: "0.8.0",
			expected:       true,
		},
		{
			name:           "current greater than minimum",
			currentVersion: "0.9.0",
			minimumVersion: "0.8.0",
			expected:       true,
		},
		{
			name:           "current less than minimum",
			currentVersion: "0.7.0",
			minimumVersion: "0.8.0",
			expected:       false,
		},
		{
			name:           "current with patch higher",
			currentVersion: "0.8.1",
			minimumVersion: "0.8.0",
			expected:       true,
		},
		{
			name:           "current with patch lower",
			currentVersion: "0.8.0",
			minimumVersion: "0.8.1",
			expected:       false,
		},
		{
			name:           "invalid current version",
			currentVersion: "invalid",
			minimumVersion: "0.8.0",
			expected:       false,
		},
		{
			name:           "invalid minimum version",
			currentVersion: "0.8.0",
			minimumVersion: "invalid",
			expected:       false,
		},
		{
			name:           "unknown current version with defined minimum",
			currentVersion: "0.0.0",
			minimumVersion: "0.8.0",
			expected:       false,
		},
		{
			name:           "with v prefix",
			currentVersion: "v0.8.0",
			minimumVersion: "0.8.0",
			expected:       true,
		},
		{
			name:           "empty minimum version",
			currentVersion: "0.8.0",
			minimumVersion: "",
			expected:       false, // Current implementation returns false for empty minimum version
		},
		{
			name:           "empty current version",
			currentVersion: "",
			minimumVersion: "0.8.0",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VersionMeetsMinimum(tt.currentVersion, tt.minimumVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSuiteWithMinimumVersion(t *testing.T) {
	tests := []struct {
		name         string
		suiteContent string
		expectSkip   bool
		skipReason   string
	}{
		{
			name: "valid minimum version",
			suiteContent: `
suite: minimumVersion test
skip:
  minimumVersion: 0.0.1
tests:
  - it: should pass
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: false,
		},
		{
			name: "minimum version too high",
			suiteContent: `
suite: minimumVersion test
skip:
  minimumVersion: 99.0.0
tests:
  - it: should fail
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: true,
			skipReason: "Test suite requires minimum unittest plugin version 99.0.0",
		},
		{
			name: "no minimum version",
			suiteContent: `
suite: minimumVersion test
tests:
  - it: should pass
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: false,
		},
		{
			name: "empty minimum version",
			suiteContent: `
suite: minimumVersion test
skip:
  minimumVersion: ""
tests:
  - it: should pass
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: false,
		},
		{
			name: "invalid version format",
			suiteContent: `
suite: minimumVersion test
skip:
  minimumVersion: "not-a-valid-version"
tests:
  - it: should pass
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: true, // Invalid version formats will also trigger skip
			skipReason: "Test suite requires minimum unittest plugin version not-a-valid-version",
		},
		{
			name: "both reason and minimumVersion specified",
			suiteContent: `
suite: minimumVersion test
skip:
  reason: "Custom skip reason"
  minimumVersion: 0.0.1
tests:
  - it: should pass
    asserts:
      - isKind:
          of: Deployment
`,
			expectSkip: true,
			skipReason: "Custom skip reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the test suite content
			file := path.Join("_scratch", fmt.Sprintf("minimum-version-%s.yaml", tt.name))
			assert.Nil(t, writeToFile(tt.suiteContent, file))
			defer os.RemoveAll(file)

			suites, err := ParseTestSuiteFile(file, "basic", true, []string{})
			assert.NoError(t, err)
			assert.NotEmpty(t, suites)

			// Check if suite was skipped correctly
			if tt.expectSkip {
				assert.NotEmpty(t, suites[0].Skip.Reason)
				assert.Contains(t, suites[0].Skip.Reason, tt.skipReason)
			} else {
				assert.Empty(t, suites[0].Skip.Reason)
			}
		})
	}
}
