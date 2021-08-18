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

func parseQueryString(original string) (parsed string, err error) {
	switch strings.ToLower(original) {
	case "last week":
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * 7)
		parsed = fmt.Sprintf(
			"date>=%s AND date<=%s", cutoff.Format("2006/01/02"),
			today.Format("2006/01/02"),
		)
		return
	case "last 30 days":
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * 30)
		parsed = fmt.Sprintf(
			"date>=%s AND date<=%s", cutoff.Format("2006/01/02"),
			today.Format("2006/01/02"),
		)
		return
	}
	parsed = original
	return
}

func initTransCmd(rootCmd *cobra.Command) {
	transLsCmd.Flags().StringP("query", "q", "", "Query string for transactions")
	transCmd.AddCommand(transLsCmd)
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
		queryStr = "last week"
	}
	queryStr, err = parseQueryString(queryStr)
	cobra.CheckErr(err)
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
