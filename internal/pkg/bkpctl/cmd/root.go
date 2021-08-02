package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bkpctl",
	Short: "The bookkeeper controller",
	Long: `The bookkeeper controller (bkpctl) provides a CLI to interact with
and manage the bookkeeper backend.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func Init() {
	initDbCmd(rootCmd)
}
