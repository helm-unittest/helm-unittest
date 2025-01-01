package main_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/helm-unittest/helm-unittest/cmd/helm-unittest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupTestCmd() *cobra.Command {
	testCmd := &cobra.Command{
		Use:           "unittest",
		Run:           RunPlugin,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// silence tests
	// redirect output to buffer
	buf := new(bytes.Buffer)
	testCmd.SetOut(buf)
	testCmd.SetErr(buf)

	// old := os.Stdout // keep backup of the real stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// back to normal state
	// w.Close()
	// out, _ := io.ReadAll(r)
	// os.Stdout = old // restoring the real stdout

	InitPluginFlags(testCmd)
	return testCmd
}

func TestValidateUnittestColorFlags(t *testing.T) {
	a := assert.New(t)

	colorFlags := map[string]bool{
		"":              false,
		"--color":       true,
		"--color=true":  true,
		"--color=false": false,
	}

	for colorFlag, colorValue := range colorFlags {
		cmd := setupTestCmd()

		// Setup actual parameter
		if len(colorFlag) > 0 {
			cmd.SetArgs([]string{colorFlag})
		} else {
			// Empty parameter to ensure no chart is found
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()
		actualResult := false // Actual default outcome depends on OS

		if runner.Printer.Colored != nil {
			actualResult = *runner.Printer.Colored
		}

		a.Nil(result)
		a.Equal(colorValue, actualResult)
	}
}

func TestValidateUnittestDebugFlags(t *testing.T) {
	a := assert.New(t)

	debugFlags := map[string]bool{
		"":                    false,
		"--debugPlugin":       true,
		"-d":                  true,
		"--debugPlugin=true":  true,
		"--debugPlugin=false": false,
	}

	for debugFlag, debugValue := range debugFlags {
		cmd := setupTestCmd()

		// Setup actual parameter
		if len(debugFlag) > 0 {
			cmd.SetArgs([]string{debugFlag})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()

		a.Nil(result)
		a.Equal(debugValue, DebugEnabled())
	}
}

func TestValidateUnittestStrictFlag(t *testing.T) {
	a := assert.New(t)

	strictFlags := map[string]bool{
		"":               false,
		"--strict":       true,
		"--strict=false": false,
		"--strict=true":  true,
	}

	for strictFlag, strictFlagValue := range strictFlags {
		cmd := setupTestCmd()

		// Setup actual parameter
		if len(strictFlag) > 0 {
			cmd.SetArgs([]string{strictFlag})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(result)
		a.Equal(strictFlagValue, runner.Strict)
	}
}

func TestValidateUnittestFailFastFlags(t *testing.T) {
	a := assert.New(t)

	failFastFlags := map[string]bool{
		"":                 false,
		"--failfast":       true,
		"-q":               true,
		"--failfast=true":  true,
		"--failfast=false": false,
	}

	for failFastFlag, failFastFlagValue := range failFastFlags {
		cmd := setupTestCmd()

		// Setup actual parameter
		if len(failFastFlag) > 0 {
			cmd.SetArgs([]string{failFastFlag})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(result)
		a.Equal(failFastFlagValue, runner.Failfast)
	}
}

func TestValidateUnittestUpdateSnapshotFlags(t *testing.T) {
	a := assert.New(t)

	updateSnapshotFlags := map[string]bool{
		"":                        false,
		"--update-snapshot":       true,
		"-u":                      true,
		"--update-snapshot=true":  true,
		"--update-snapshot=false": false,
	}

	for updateSnapshotFlag, updateSnapshotFlagValue := range updateSnapshotFlags {
		cmd := setupTestCmd()
		// Setup actual parameter
		if len(updateSnapshotFlag) > 0 {
			cmd.SetArgs([]string{updateSnapshotFlag})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(result)
		a.Equal(updateSnapshotFlagValue, runner.UpdateSnapshot)
	}
}

func TestValidateUnittestWithSnapshotFlags(t *testing.T) {
	a := assert.New(t)

	withSubchartFlags := map[string]bool{
		"":                      true,
		"--with-subchart":       true,
		"-u":                    true,
		"--with-subchart=true":  true,
		"--with-subchart=false": false,
	}

	for withSubchartFlag, withSubchartFlagValue := range withSubchartFlags {
		cmd := setupTestCmd()
		// Setup actual parameter
		if len(withSubchartFlag) > 0 {
			cmd.SetArgs([]string{withSubchartFlag})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(result)
		a.Equal(withSubchartFlagValue, runner.WithSubChart)
	}
}

func TestValidateUnittestTestFilesFlags(t *testing.T) {
	a := assert.New(t)

	testFileFlags := []string{"--file", "-f"}

	testFiles := map[string][]string{
		"":             {filepath.Join("tests", "*_test.yaml")},
		"*.yaml":       {"*.yaml"},
		"*_tests.yaml": {"*_tests.yaml"},
	}

	for _, testFileFlag := range testFileFlags {
		for testFile, testFileValues := range testFiles {
			cmd := setupTestCmd()
			if len(testFile) > 0 {
				cmd.SetArgs([]string{testFileFlag, testFile})
			} else {
				cmd.SetArgs([]string{})
			}

			result := cmd.Execute()
			runner := GetTestRunner()

			a.Nil(result)
			a.EqualValues(testFileValues, runner.TestFiles)
		}
	}
}

// values
func TestValidateUnittestValuesFlags(t *testing.T) {
	a := assert.New(t)

	valuesFilesFlags := []string{"--values", "-v"}

	valuesFiles := map[string][]string{
		"":              {},
		"*_values.yaml": {"*_values.yaml"},
		"values.yaml":   {"values.yaml"},
	}

	for _, valuesFilesFlag := range valuesFilesFlags {
		for valuesFile, valuesFileValues := range valuesFiles {
			cmd := setupTestCmd()
			if len(valuesFile) > 0 {
				cmd.SetArgs([]string{valuesFilesFlag, valuesFile})
			} else {
				cmd.SetArgs([]string{})
			}

			result := cmd.Execute()
			runner := GetTestRunner()

			a.Nil(result)
			a.EqualValues(valuesFileValues, runner.ValuesFiles)
		}
	}
}

// output-file
func TestValidateUnittestOutputFileFlags(t *testing.T) {
	a := assert.New(t)

	outputFileFlags := []string{"--output-file", "-o"}

	outputFiles := map[string]string{
		"":                "",
		"test-output.xml": "test-output.xml",
	}

	for _, outputFileFlag := range outputFileFlags {
		for outputFile, outputFileValue := range outputFiles {
			defer os.Remove(outputFile)
			cmd := setupTestCmd()
			if len(outputFile) > 0 {
				cmd.SetArgs([]string{outputFileFlag, outputFile})
			} else {
				cmd.SetArgs([]string{})
			}

			result := cmd.Execute()
			runner := GetTestRunner()

			a.Nil(result)
			a.EqualValues(outputFileValue, runner.OutputFile)
		}
	}
}

// output-type
func TestValidateUnittestOutputTypeFlags(t *testing.T) {
	a := assert.New(t)

	dummyOutputFile := "test-output.xml"
	defer os.Remove(dummyOutputFile)
	outputTypeFlags := []string{"--output-type", "-t"}

	outputTypes := map[string]string{
		"":      "*formatter.xUnitReportXML",
		"JUnit": "*formatter.jUnitReportXML",
		"NUnit": "*formatter.nUnitReportXML",
		"XUnit": "*formatter.xUnitReportXML",
		"Sonar": "*formatter.sonarReportXML",
	}

	for _, outputTypeFlag := range outputTypeFlags {
		for outputType, outputTypeValue := range outputTypes {
			cmd := setupTestCmd()
			if len(outputType) > 0 {
				cmd.SetArgs([]string{"-o", dummyOutputFile, outputTypeFlag, outputType})
			} else {
				cmd.SetArgs([]string{"-o", dummyOutputFile})
			}

			result := cmd.Execute()
			runner := GetTestRunner()

			a.Nil(result)
			a.Equal(outputTypeValue, typeofObject(runner.Formatter))
		}
	}
}

// chart-test-path
func TestValidateUnittestChartTestsPathFlag(t *testing.T) {
	a := assert.New(t)

	chartTestPathFlag := "--chart-tests-path"

	chartTestPaths := map[string]string{
		"":                 "",
		".":                ".",
		"../central-tests": "../central-tests",
	}

	for chartTestPath, chartTestPathValue := range chartTestPaths {
		cmd := setupTestCmd()
		if len(chartTestPath) > 0 {
			cmd.SetArgs([]string{chartTestPathFlag, chartTestPath})
		} else {
			cmd.SetArgs([]string{})
		}

		result := cmd.Execute()
		runner := GetTestRunner()

		a.Nil(result)
		a.EqualValues(chartTestPathValue, runner.ChartTestsPath)
	}
}

// Using %T
func typeofObject(variable interface{}) string {
	return fmt.Sprintf("%T", variable)
}
