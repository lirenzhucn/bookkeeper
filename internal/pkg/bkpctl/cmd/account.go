package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/lensesio/tableprinter"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Create, query, and update account info via API",
}
var accountLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all accounts",
	Run:   lsAccounts,
}

func initAccountCmd(rootCmd *cobra.Command) {
	accountCmd.AddCommand(accountLsCmd)
	rootCmd.AddCommand(accountCmd)
}

func lsAccounts(cmd *cobra.Command, args []string) {
	var err error
	var accounts []bookkeeper.Account
	resp, err := http.Get(BASE_URL + "accounts")
	cobra.CheckErr(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	cobra.CheckErr(err)
	err = json.Unmarshal(body, &accounts)
	cobra.CheckErr(err)
	printer := tableprinter.New(os.Stdout)
	printer.Print(accounts)
}
