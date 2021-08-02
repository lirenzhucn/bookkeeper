package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kataras/tablewriter"
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
	tablePrintAccounts(accounts)
}

func tablePrintAccounts(accounts []bookkeeper.Account) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Desc", "Tags"})
	for _, a := range accounts {
		row := []string{fmt.Sprintf("%d", a.Id), a.Name, a.Desc, strings.Join(a.Tags, ", ")}
		table.Append(row)
	}
	table.Render()
}
