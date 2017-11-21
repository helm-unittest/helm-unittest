package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := &cobra.Command{
		Use:   "test [flags] CHART",
		Short: "test the helm chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("hello helm test!")
			return nil
		},
	}
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
