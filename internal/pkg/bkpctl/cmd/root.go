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

var BASE_URL = "http://localhost:10000/"

func Execute() error {
	return rootCmd.Execute()
}

func Init() {
	initDbCmd(rootCmd)
	initAccountCmd(rootCmd)
	initTransCmd(rootCmd)
}
