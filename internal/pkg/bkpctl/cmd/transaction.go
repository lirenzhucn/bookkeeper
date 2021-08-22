package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/olekukonko/tablewriter"
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
var transUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a transaction",
	Run:   updateTransactions,
}
var transReconCmd = &cobra.Command{
	Use:   "recon",
	Short: "Reconcile one account",
	Run: func(cmd *cobra.Command, args []string) {
		dateRange_, err := cmd.Flags().GetString("date-range")
		cobra.CheckErr(err)
		if dateRange_ == "" {
			dateRange_ = "past week"
		}
		dateRange, ok := parseDateRange(dateRange_)
		if !ok {
			cobra.CheckErr(fmt.Errorf("invalid date range %s", dateRange_))
		}
		accountName, err := cmd.Flags().GetString("account")
		cobra.CheckErr(err)
		account, err := getAccountByName(accountName)
		cobra.CheckErr(err)
		fmt.Println(dateRange)
		fmt.Println(account)
	},
}

func parseQueryString(original string) (parsed string) {
	phrases := strings.SplitN(original, "on", 2)
	if len(phrases) < 2 {
		phrases = append(phrases, "")
	}
	dateRange, accountName := phrases[0], phrases[1]
	dateRange = strings.TrimSpace(dateRange)
	accountName = strings.TrimSpace(accountName)
	parsedDateRange, ok := dateRangeToQuery(dateRange)
	if ok {
		parsed = parsed + parsedDateRange
	}
	if accountName != "" {
		parsed = parsed + fmt.Sprintf(` AND a.name="%s"`, accountName)
	}
	return
}

func initTransCmd(rootCmd *cobra.Command) {
	transLsCmd.Flags().StringP("query", "q", "", "Query string for transactions")
	transUpdateCmd.Flags().StringP(
		"categories", "c", "",
		"Path to the Category definition file (default: ./configs/category_map.json)",
	)
	transReconCmd.Flags().StringP(
		"account", "a", "",
		"The name of the account to reconcile",
	)
	transReconCmd.Flags().StringP(
		"date-range", "d", "",
		"The date range to reconcile (default: past week)",
	)
	transReconCmd.MarkFlagRequired("account")
	transCmd.AddCommand(transLsCmd)
	transCmd.AddCommand(transUpdateCmd)
	transCmd.AddCommand(transReconCmd)
	rootCmd.AddCommand(transCmd)
}

func lsTransactions(cmd *cobra.Command, args []string) {
	var (
		err      error
		queryStr string
		url_     string
	)

	queryStr, err = cmd.Flags().GetString("query")
	cobra.CheckErr(err)
	if queryStr == "" {
		queryStr = "past week"
	}
	queryStr = parseQueryString(queryStr)
	url_ = fmt.Sprintf("%stransactions?queryString=%s", BASE_URL,
		url.QueryEscape(queryStr))

	var transactions []bookkeeper.Transaction_
	resp, err := http.Get(url_)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	cobra.CheckErr(err)
	err = json.Unmarshal(body, &transactions)
	cobra.CheckErr(err)
	tablePrintTransactions(transactions)
}

func tablePrintTransactions(transactions []bookkeeper.Transaction_) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Id", "Type", "Date", "Category", "Sub-Category", "Account Name",
		"Amount", "Notes", "Association Id",
	})
	for _, t := range transactions {
		row := []string{
			fmt.Sprintf("%d", t.Id), t.Type, t.Date.Format("2006/01/02"),
			t.Category, t.SubCategory, t.AccountName,
			fmt.Sprintf("%.2f", float32(t.Amount)/100.0), t.Notes,
			t.AssociationId,
		}
		table.Append(row)
	}
	table.Render()
}

func updateTransactions(cmd *cobra.Command, args []string) {
	// read category map
	categoriesFile, err := cmd.Flags().GetString("categories")
	cobra.CheckErr(err)
	var categoryMap CategoryMap
	readCategoryMap(categoriesFile, &categoryMap)
	// get all accounts
	var accounts []bookkeeper.Account
	getAllAccounts(&accounts)

	var entry JournalEntry
	err = entry.InteractiveSingleUpdate(accounts, categoryMap)
	cobra.CheckErr(err)
	if entry.InteractiveConfirm() {
		fmt.Println("Updating the journal entry to the server...")
		err = entry.PatchToServer()
		cobra.CheckErr(err)
	} else {
		fmt.Println("No journal entries or transactions are updated.")
	}
}
