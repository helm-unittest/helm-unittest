package helmtest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type TestConfig struct {
	Colored   bool
	TestFiles []string
}

var testConfig = TestConfig{}

var cmd = &cobra.Command{
	Use:   "unittest [flags] CHART [...]",
	Short: "unittest for helm charts",
	Long: `This helps you
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

var Colored bool

func init() {
	cmd.PersistentFlags().BoolVar(
		&testConfig.Colored, "color", false,
		"enforce printing colored output even stdout is not a tty. Set to false to disable color.",
	)
	defaultFilePattern := filepath.Join("tests", "*_test.yaml")
	cmd.PersistentFlags().StringArrayVarP(
		&testConfig.TestFiles, "file", "f", []string{defaultFilePattern},
		"glob paths of test files location, default to "+defaultFilePattern,
	)
}
