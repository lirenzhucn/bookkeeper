package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kataras/tablewriter"
	"github.com/leekchan/accounting"
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
var accountBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show account balance at a given date",
	Args:  cobra.NoArgs,
	Run:   accountBalance,
}

func initAccountCmd(rootCmd *cobra.Command) {
	accountLsCmd.Flags().IntP("id", "i", -1, "specify an specific id to list")
	accountBalanceCmd.Flags().StringP("name", "n", "", "specify the account name")
	accountBalanceCmd.Flags().StringP(
		"date", "d", "", "specify the date (default: today local time)")
	accountCmd.AddCommand(accountLsCmd)
	accountCmd.AddCommand(accountBalanceCmd)
	rootCmd.AddCommand(accountCmd)
}

func accountBalance(cmd *cobra.Command, args []string) {
	name, err := cmd.Flags().GetString("name")
	cobra.CheckErr(err)
	date, _ := cmd.Flags().GetString("date")
	if date == "" {
		date = time.Now().Format("2006/01/02")
	}
	if name == "" {
		accounts_, err := allAccountsBalance(date)
		cobra.CheckErr(err)
		tablePrintAccountsWithBalance(accounts_)
	} else {
		account_, err := singleAccountBalance(name, date)
		cobra.CheckErr(err)
		tablePrintAccountsWithBalance([]bookkeeper.AccountWithBalance{account_})
	}
}

func tablePrintAccountsWithBalance(accounts_ []bookkeeper.AccountWithBalance) {
	// table print
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Desc", "Tags", "Balance"})
	for _, account_ := range accounts_ {
		table.Append([]string{
			fmt.Sprintf("%d", account_.Id), account_.Name, account_.Desc,
			strings.Join(account_.Tags, ", "),
			ac.FormatMoney(float64(account_.Balance) / 100),
		})
	}
	table.Render()
}

func allAccountsBalance(date string) ([]bookkeeper.AccountWithBalance, error) {
	var accounts_ []bookkeeper.AccountWithBalance
	url_ := fmt.Sprintf("%sreporting/account_balance?date=%s",
		BASE_URL, url.QueryEscape(date))
	resp, err := http.Get(url_)
	if err != nil {
		return accounts_, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&accounts_)
	return accounts_, nil
}

func singleAccountBalance(
	name string, date string) (bookkeeper.AccountWithBalance, error) {
	var account_ bookkeeper.AccountWithBalance
	url_ := fmt.Sprintf("%sreporting/account_balance?accountName=%s&date=%s",
		BASE_URL, url.QueryEscape(name), url.QueryEscape(date))
	resp, err := http.Get(url_)
	if err != nil {
		return account_, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf(
			"Failed to get the balance for account %s (response status: %s)\n",
			name, resp.Status)
		return account_, fmt.Errorf(
			"failed to get the balance for account %s (response status: %s)",
			name, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return account_, err
	}
	err = json.Unmarshal(body, &account_)
	if err != nil {
		return account_, err
	}
	return account_, nil
}

func lsAccounts(cmd *cobra.Command, args []string) {
	var (
		err      error
		accounts []bookkeeper.Account
		id       int
		url_     string
	)
	singleAccount := false

	url_ = BASE_URL + "accounts"
	id, err = cmd.Flags().GetInt("id")
	if err == nil && id >= 0 {
		url_ += fmt.Sprintf("/%d", id)
		singleAccount = true
	}

	resp, err := http.Get(url_)
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
