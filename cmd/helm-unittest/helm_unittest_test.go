package main_test

import (
	"testing"

	. "github.com/helm-unittest/helm-unittest/cmd/helm-unittest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var existingLongNameFlags = []string{
	"chart-tests-path",
	"color",
	"debug",
	"failfast",
	"file",
	"output-file",
	"output-type",
	"strict",
	"values",
	"update-snapshot",
	"with-subchart",
}

func setupTestCmd() *cobra.Command {
	testCmd := &cobra.Command{
		Use: "unittest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	InitPluginFlags(testCmd)
	return testCmd
}

func setupRootTestCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "helm-unittest",
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "helm unittest",
		},
	}
	rootCmd.PersistentFlags().Bool("debug", false, "Parent boolean")
	rootCmd.AddCommand(setupTestCmd())
	return rootCmd
}

func TestValidateUnittestLongNameFlags(t *testing.T) {
	cmd := setupTestCmd()

	for _, flag := range existingLongNameFlags {
		foundFlag := cmd.Flag(flag)
		assert.NotNil(t, foundFlag, flag)
	}
}

func TestValidateUnittestInheritedFlag(t *testing.T) {
	cmd := setupRootTestCmd()

	inheritedFlag := "debug"

	foundFlag := cmd.Flag(inheritedFlag)
	assert.NotNil(t, foundFlag, inheritedFlag)
}
