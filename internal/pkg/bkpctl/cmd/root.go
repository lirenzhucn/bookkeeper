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

const BASE_URL string = "http://localhost:10000/"
const BKPCTL_DATE_FORMAT string = "2006/01/02"

func Execute() error {
	return rootCmd.Execute()
}

func Init() {
	initDbCmd(rootCmd)
	initImportCmd(rootCmd)
	initAccountCmd(rootCmd)
	initTransCmd(rootCmd)
	initReportCmd(rootCmd)
	initRecordCmd(rootCmd)
}
