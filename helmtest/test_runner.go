package helmtest

import (
	"fmt"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
)

func getTestSuiteFiles(chartPath string) ([]string, error) {
	return filepath.Glob(filepath.Join(chartPath, "tests", "*.yaml"))
}

type TestRunner struct {
	ChartsPath []string
}

func (tr TestRunner) Run(logger loggable) bool {
	allPassed := true
	for _, chartPath := range tr.ChartsPath {
		chart, err := chartutil.Load(chartPath)
		if err != nil {
			fmt.Print(err)
		}

		suiteFiles, err := getTestSuiteFiles(chartPath)
		if err != nil {
			fmt.Print(err)
		}

		suitesResult := make([]*TestSuiteResult, len(suiteFiles))
		for idx, file := range suiteFiles {
			testSuite, err := ParseTestSuiteFile(file)
			if err != nil {
				suitesResult[idx] = &TestSuiteResult{
					FilePath:  file,
					ExecError: err,
				}
			}
			result := testSuite.Run(chart, &TestSuiteResult{FilePath: file})
			allPassed = allPassed && result.Passed
			suitesResult[idx] = result
		}

		for _, result := range suitesResult {
			result.print(logger, 0)
		}
	}
	return false
}
