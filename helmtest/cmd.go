package helmtest

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "unittest [flags] CHART [...]",
	Short: "unittest for helm charts",
	Long: `This helps you
  `,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, chartsPath []string) {
		runner := TestRunner{ChartsPath: chartsPath}
		printer := Printer{Writer: os.Stdout, Colored: true}
		passed := runner.Run(&printer)

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
