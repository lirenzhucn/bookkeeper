package cmd

import (
	"fmt"
	"time"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate various financial reports",
}
var balanceSheetCmd = &cobra.Command{
	Use:   "balance-sheet",
	Short: "Generate the balance sheet of a specified date",
	Args:  cobra.NoArgs,
	Run:   generateBalanceSheet,
}

func initReportCmd(rootCmd *cobra.Command) {
	balanceSheetCmd.Flags().StringSliceP(
		"asset-tags", "a",
		[]string{"cash", "taxable+liquid", "retirement", "education",
			"nonliquid", "real estate"},
		"Specify a list of asset tags to report separately",
	)
	balanceSheetCmd.Flags().StringSliceP(
		"liability-tags", "l",
		[]string{"credit card", "loan"},
		"Specify a list of liability tags to report separately",
	)
	balanceSheetCmd.Flags().StringP("date", "d", "",
		"Specify the date to create the balance sheet for (default: today)")
	reportCmd.AddCommand(balanceSheetCmd)
	rootCmd.AddCommand(reportCmd)
}

func generateBalanceSheet(cmd *cobra.Command, args []string) {
	assetTags, err := cmd.Flags().GetStringSlice("asset-tags")
	cobra.CheckErr(err)
	liabilityTags, err := cmd.Flags().GetStringSlice("liability-tags")
	cobra.CheckErr(err)
	dateStr, _ := cmd.Flags().GetString("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006/01/02")
	}

	accounts_, err := allAccountsBalance(dateStr)
	cobra.CheckErr(err)

	balanceSheet := bookkeeper.ComputeBalanceSheet(accounts_, assetTags, liabilityTags)
	fmt.Println(balanceSheet)
}
