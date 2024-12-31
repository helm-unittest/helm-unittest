package unittest_test

import (
	"testing"

	. "github.com/helm-unittest/helm-unittest/pkg/unittest"
	"github.com/helm-unittest/helm-unittest/pkg/unittest/results"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func TestV3RunSkippedJob(t *testing.T) {
	c, _ := loader.Load(testV3BasicChart)
	manifest := `
it: should skip when not supported assert is found
template: templates/deployment.yaml
documentIndex: 0
skip:
 reason: "Skip this test"
asserts:
  - equal:
      path: kind
      value: Deployment
`
	var tj TestJob
	_ = yaml.Unmarshal([]byte(manifest), &tj)

	testResult := tj.RunV3(c, nil, true, "", &results.TestJobResult{})
	assert.True(t, testResult.Skipped)
}
