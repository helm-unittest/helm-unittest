package unittest_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

// Most used test files
const testSuiteTests string = "_suite_tests"

const testValuesFiles = "../../test/data/services_values.yaml"
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
const testV3GlobalDoubleChart string = "../../test/data/v3/global-double-setting"
const testV3WithHelmTestsChart string = "../../test/data/v3/with-helm-tests"

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
func getExpectedRenderedTestSuites(customSnapshotIds bool) map[string]*TestSuite {
	// multiple_suites_snapshot.yaml assertions
	createSnapshotTestYaml := func(env string)string {
		return fmt.Sprintf(`
it: manifest should match snapshot
set: 
    env: %s
asserts:
    - matchSnapshot: {}`, env)
	}
	snapshotDevTest := TestJob{}
	_ = yaml.Unmarshal([]byte(createSnapshotTestYaml("dev")), &snapshotDevTest)
	snapshotProdTest := TestJob{}
	_ = yaml.Unmarshal([]byte(createSnapshotTestYaml("prod")), &snapshotProdTest)
	// multiple_test_suites.yaml assertions
	crateMultipleTestSuitesYaml := func(env string)string {
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
	_ = yaml.Unmarshal([]byte(crateMultipleTestSuitesYaml("dev")), &multipleTestSuitesDevTest)
	multipleTestSuitesProdTest := TestJob{}
	_ = yaml.Unmarshal([]byte(crateMultipleTestSuitesYaml("prod")), &multipleTestSuitesProdTest)
	// multiple_tests_test.yaml assertions
	var secretNameEqualsYaml = func(env string)string {
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
	_ = yaml.Unmarshal([]byte(secretNameEqualsYaml("dev")), &multipleTestsDevTest)
	multipleTestsProdTest := TestJob{}
	_ = yaml.Unmarshal([]byte(secretNameEqualsYaml("prod")), &multipleTestsProdTest)
	const multipleTestsFirstTestYaml = `
it: should render nothing if not enabled
asserts:
    - hasDocuments:
        count: 0`
	multipleTestsFirstTest := TestJob{}
	_ = yaml.Unmarshal([]byte(multipleTestsFirstTestYaml), &multipleTestsFirstTest)

	// Set up snapshotId values
	// Note, this is completely based on the order of the yaml in a single suite template file
	var (
		multipleTestSuiteDevSnapshotId string
		multipleTestSuiteProdSnapshotId string
		multipleSuiteSnapshotsDevSnapshotId string
		multipleSuiteSnapshotsProdSnapshotId string
		multipleTestsSnapshotId string
	)
	if (customSnapshotIds) {
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

	return map[string]*TestSuite {
		"multiple test suites dev": {
			Templates: []string{"charts/postgresql/templates/secrets.yaml"},
			SnapshotId: multipleTestSuiteDevSnapshotId,
			Tests: []*TestJob{
				&multipleTestSuitesDevTest,
			},
		},
		"multiple test suites prod": {
			Templates: []string{"charts/postgresql/templates/secrets.yaml"},
			SnapshotId: multipleTestSuiteProdSnapshotId,
			Tests: []*TestJob {
				&multipleTestSuitesProdTest,
			},
		},
		"multiple test suites snapshot dev": {
			Templates: []string{"templates/service.yaml"},
			SnapshotId: multipleSuiteSnapshotsDevSnapshotId,
			Tests: []*TestJob {
				&snapshotDevTest,
			},
		},
		"multiple test suites snapshot prod": {
			Templates: []string{"templates/service.yaml"},
			SnapshotId: multipleSuiteSnapshotsProdSnapshotId,
			Tests: []*TestJob {
				&snapshotProdTest,
			},
		},
		"multiple tests": {
			Templates: []string{"templates/ingress.yaml"},
			SnapshotId: multipleTestsSnapshotId,
			Tests: []*TestJob {
				&multipleTestsFirstTest,
				&multipleTestsDevTest,
				&multipleTestsProdTest,
			},
		},
	}
}

func TestV3ParseTestSuiteUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_test.yaml", "basic", false, []string{})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV3ParseTestSuiteUnstrictNoTestsFileFail(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_notests_test.yaml", "basic", false, []string{})

	a.NotNil(err)
	a.EqualError(err, "no tests found")
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
}

func TestV3ParseTestSuiteUnstrictNoAssertsFileFail(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_noasserts_test.yaml", "basic", false, []string{})

	a.NotNil(err)
	a.EqualError(err, "no asserts found")
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV3ParseTestSuiteStrictFileError(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/invalidbasic/tests/deployment_test.yaml", "basic", true, []string{})

	a.NotNil(err)
	a.EqualError(err, "yaml: unmarshal errors:\n  line 6: field documents not found in type unittest.TestJob")
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"templates/deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
}

func TestV3ParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{})

	a.Nil(err)
	a.Equal(suite.Name, "test deployment")
	a.Equal(suite.Templates, []string{"templates/configmap.yaml", "templates/deployment.yaml"})
	a.Equal(suite.Tests[0].Name, "should pass all kinds of assertion")
}

func TestV3ParseTestSuiteFileWithOverrideValuesOk(t *testing.T) {
	a := assert.New(t)
	suite, err := ParseTestSuiteFile("../../test/data/v3/basic/tests/deployment_test.yaml", "basic", true, []string{testValuesFiles})

	a.Nil(err)
	a.Equal("test deployment", suite.Name)
	a.Equal([]string{"templates/configmap.yaml", "templates/deployment.yaml"}, suite.Templates)
	a.Equal("should pass all kinds of assertion", suite.Tests[0].Name)
	a.Equal(1, len(suite.Values)) // Expect services_values.yaml
}

func TestV3RenderSuitesUnstrictFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", false, []string{}, map[string]interface{} {
		"unexpectedField": false,
	})

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(false)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite" + suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name (" + suite.Name + ") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name (" + suite.Name + ") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name (" + suite.Name + ") mismatched tests")
	}
}

func TestV3RenderSuitesStrictFileFail(t *testing.T) {
	a := assert.New(t)
	_, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{} {
		"unexpectedField": true,
	})

	a.NotNil(err)
	a.ErrorContains(err, "field something not found in type unittest.TestSuite")
}

func TestV3RenderSuitesFailNoSuiteName(t *testing.T) {
	a := assert.New(t)
	_, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{} {
		"includeSuite": false,
	})

	a.NotNil(err)
	a.ErrorContains(err, "helm chart based test suites must include `suite` field")
}

func TestV3RenderSuitesStrictFileOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, nil)

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(false)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite" + suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name (" + suite.Name + ") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name (" + suite.Name + ") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name (" + suite.Name + ") mismatched tests")
	}
}

func TestV3RenderSuitesCustomSnapshotIdOk(t *testing.T) {
	a := assert.New(t)
	suites, err := RenderTestSuiteFiles("../../test/data/v3/with-helm-tests/tests-chart", "basic", true, []string{}, map[string]interface{} {
		"customSnapshotIds": true,
	})

	a.Nil(err)

	expectedSuites := getExpectedRenderedTestSuites(true)

	for _, suite := range suites {
		a.Contains(expectedSuites, suite.Name, "Unexpected test suite" + suite.Name)
		expected := expectedSuites[suite.Name]
		a.EqualValues(expected.Templates, suite.Templates, "Suite Name (" + suite.Name + ") mismatched templates")
		a.Equal(expected.SnapshotId, suite.SnapshotId, "Suite Name (" + suite.Name + ") unexpected Snapshot Id")
		a.EqualValues(expected.Tests, suite.Tests, "Suite Name (" + suite.Name + ") mismatched tests")
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
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "v3_noasserts_template_test.yaml"), false)
	suiteResult := testSuite.RunV3(testV3BasicChart, cache, true, &results.TestSuiteResult{})

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
