package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/leekchan/accounting"
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
		// get account balance at start date
		oneDay, _ := time.ParseDuration("24h")
		account, err := singleAccountBalance(
			accountName, dateRange.startDate.Add(-oneDay).Format("2006/01/02"))
		cobra.CheckErr(err)
		// get transactions between
		queryStr := fmt.Sprintf(
			`date>=%s AND date<=%s AND a.name="%s"`,
			dateRange.startDate.Format("2006/01/02"),
			dateRange.endDate.Format("2006/01/02"),
			accountName,
		)
		transactions, err := getTransactionsByQuery(queryStr)
		cobra.CheckErr(err)
		tablePrintReconcile(transactions, account.Balance)
	},
}
var transDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete one transaction",
	Run: func(cmd *cobra.Command, args []string) {
		transId, err := cmd.Flags().GetInt("id")
		cobra.CheckErr(err)
		yes, err := cmd.Flags().GetBool("yes")
		cobra.CheckErr(err)
		var trans bookkeeper.Transaction_
		err = getTransactionById(transId, &trans)
		cobra.CheckErr(err)
		tablePrintTransactions([]bookkeeper.Transaction_{trans})
		if !yes {
			survey.AskOne(&survey.Confirm{
				Message: "Are you sure that you want to delete this transaction?",
			}, &yes)
		}
		if yes {
			err = deleteSingleTransaction(transId)
			cobra.CheckErr(err)
		}
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
	transDeleteCmd.Flags().IntP("id", "i", -1, "ID of the transaction to delete")
	transDeleteCmd.MarkFlagRequired("id")
	transDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation if set")
	transReconCmd.MarkFlagRequired("account")
	transCmd.AddCommand(transLsCmd)
	transCmd.AddCommand(transUpdateCmd)
	transCmd.AddCommand(transReconCmd)
	transCmd.AddCommand(transDeleteCmd)
	rootCmd.AddCommand(transCmd)
}

func lsTransactions(cmd *cobra.Command, args []string) {
	var (
		err      error
		queryStr string
	)

	queryStr, err = cmd.Flags().GetString("query")
	cobra.CheckErr(err)
	if queryStr == "" {
		queryStr = "past week"
	}
	queryStr = parseQueryString(queryStr)
	transactions, err := getTransactionsByQuery(queryStr)
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

func tablePrintReconcile(transactions []bookkeeper.Transaction_, startingBalance int64) {
	// sort transactions
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.After(transactions[j].Date)
	})
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Id", "Type", "Date", "Category", "Sub-Category", "Account Name",
		"Amount", "Balance", "Notes",
	})
	balance := startingBalance
	for _, t := range transactions {
		balance += t.Amount
	}
	for _, t := range transactions {
		row := []string{
			fmt.Sprintf("%d", t.Id), t.Type, t.Date.Format("2006/01/02"),
			t.Category, t.SubCategory, t.AccountName,
			ac.FormatMoney(float32(t.Amount) / 100.0),
			ac.FormatMoney(float32(balance) / 100.0),
			t.Notes,
		}
		table.Append(row)
		balance -= t.Amount
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
