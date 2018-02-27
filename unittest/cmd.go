package unittest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type TestConfig struct {
	Colored        bool
	UpdateSnapshot bool
	TestFiles      []string
}

var testConfig = TestConfig{}

var cmd = &cobra.Command{
	Use:   "unittest [flags] CHART [...]",
	Short: "unittest for helm charts",
	Long: `Running chart unittest written in YAML.

This renders your chart locally (without tiller) and runs tests
defined in test suite files. Simplest test suite file looks like
below:

---
# CHART_PATH/tests/deployment_test.yaml
suite: test my deployment
templates:
  - deployment.yaml
tests:
  - it: should be a Deployment
    asserts:
      - isKind:
          of: Deployment
---

Put the test files in "tests" directory under your chart with
suffix "_test.yaml", and run:

$ helm unittest my-chart

Or specify the suite files glob path pattern:

$ helm unittest -f 'my-tests/*.yaml' my-chart

Check https://github.com/lrills/helm-unittest for more detail
about how to write tests.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, chartsPath []string) {
		if cmd.PersistentFlags().Changed("color") {
			color.NoColor = !testConfig.Colored
		}
		runner := TestRunner{ChartsPath: chartsPath}
		printer := Printer{Writer: os.Stdout, Colored: !color.NoColor}
		passed := runner.Run(&printer, testConfig)

		if !passed {
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cmd.PersistentFlags().BoolVar(
		&testConfig.Colored, "color", false,
		"enforce printing colored output even stdout is not a tty. Set to false to disable color",
	)

	defaultFilePattern := filepath.Join("tests", "*_test.yaml")
	cmd.PersistentFlags().StringArrayVarP(
		&testConfig.TestFiles, "file", "f", []string{defaultFilePattern},
		"glob paths of test files location, default to "+defaultFilePattern,
	)

	cmd.PersistentFlags().BoolVarP(
		&testConfig.UpdateSnapshot, "update-snapshot", "u", false,
		"update the snapshot cached if needed, make sure you review the change before update",
	)
}
