package unittest_test

import (
	"io/ioutil"
	"path"
	"testing"

	. "github.com/lrills/helm-unittest/unittest"
	"github.com/lrills/helm-unittest/unittest/snapshot"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
)

var tmpdir, _ = ioutil.TempDir("", "_suite_tests")

func TestParseTestSuiteFileOk(t *testing.T) {
	a := assert.New(t)
	s, err := ParseTestSuiteFile("../__fixtures__/basic/tests/deployment_test.yaml")

	a.Nil(err)
	a.Equal(s.Name, "test deployment")
	a.Equal(s.Templates, []string{"deployment.yaml"})
	a.Equal(s.Tests[0].Name, "should pass all kinds of assertion")
}

func TestRunSuiteWhenPass(t *testing.T) {
	c, _ := chartutil.Load("../__fixtures__/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
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

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.Run(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	a.True(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)

	a.Equal(uint(1), suiteResult.SnapshotCounting.Created)
	a.Equal(uint(1), suiteResult.SnapshotCounting.Total)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Failed)
	a.Equal(uint(0), suiteResult.SnapshotCounting.Vanished)
}

func TestRunSuiteWhenFail(t *testing.T) {
	c, _ := chartutil.Load("../__fixtures__/basic")
	suiteDoc := `
suite: test suite name
templates:
  - deployment.yaml
tests:
  - it: should fail
    asserts:
      - equal:
          path: kind
          value: Pod
`
	testSuite := TestSuite{}
	yaml.Unmarshal([]byte(suiteDoc), &testSuite)

	cache, _ := snapshot.CreateSnapshotOfSuite(path.Join(tmpdir, "my_test.yaml"), false)
	suiteResult := testSuite.Run(c, cache, &TestSuiteResult{})

	a := assert.New(t)
	a.False(suiteResult.Passed)
	a.Nil(suiteResult.ExecError)
	a.Equal(1, len(suiteResult.TestsResult))
	a.Equal("test suite name", suiteResult.DisplayName)
}
