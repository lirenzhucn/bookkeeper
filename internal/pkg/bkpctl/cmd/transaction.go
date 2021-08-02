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

var transCmd = &cobra.Command{
	Use:   "trans",
	Short: "Create, query, and update transactions via API",
}
var transLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List transactions",
	Long:  "List either all transactions or those between two dates",
	Run:   lsTransactions,
}

func initTransCmd(rootCmd *cobra.Command) {
	transCmd.AddCommand(transLsCmd)
	rootCmd.AddCommand(transCmd)
}

func lsTransactions(cmd *cobra.Command, args []string) {
	var err error
	var transactions []bookkeeper.Transaction
	resp, err := http.Get(BASE_URL + "transactions")
	cobra.CheckErr(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	cobra.CheckErr(err)
	err = json.Unmarshal(body, &transactions)
	cobra.CheckErr(err)
	printer := tableprinter.New(os.Stdout)
	printer.Print(transactions)
}
