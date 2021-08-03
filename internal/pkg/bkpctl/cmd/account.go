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
	accountLsCmd.Flags().IntP("id", "i", -1, "specify an specific id to list")
	accountCmd.AddCommand(accountLsCmd)
	rootCmd.AddCommand(accountCmd)
}

func lsAccounts(cmd *cobra.Command, args []string) {
	var (
		err      error
		accounts []bookkeeper.Account
		id       int
		url      string
	)
	singleAccount := false

	url = BASE_URL + "accounts"
	id, err = cmd.Flags().GetInt("id")
	if err == nil && id >= 0 {
		url += fmt.Sprintf("/%d", id)
		singleAccount = true
	}

	resp, err := http.Get(url)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf(
			"Failed to get the account(s). Response status: %s\n", resp.Status,
		)
		return
	}
	body, err := io.ReadAll(resp.Body)
	cobra.CheckErr(err)
	if singleAccount {
		body = append([]byte("["), body...)
		body = append(body, []byte("]")...)
	}
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
