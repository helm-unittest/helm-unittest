package unittest_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/snapshot"
	"github.com/stretchr/testify/assert"
)

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
`

	a := assert.New(t)
	file := path.Join("_scratch", "multiple-suites-with-skip.yaml")
	a.Nil(writeToFile(suiteDoc, file))
	defer os.RemoveAll(file)

	suites, err := ParseTestSuiteFile(file, "basic", true, []string{})

	assert.NoError(t, err)
	assert.Len(t, suites, 4)

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
		default:
			assert.Empty(t, s.Skip.Reason)
		}
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
			Reason string `yaml:"reason"`
		}{Reason: "skip suite"},
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
