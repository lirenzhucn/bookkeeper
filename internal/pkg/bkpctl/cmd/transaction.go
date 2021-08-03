package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kataras/tablewriter"
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
	transLsCmd.Flags().StringP("start", "s", "", "Start date of the query YYYY/MM/DD")
	transLsCmd.Flags().StringP("end", "e", "", "End date of the query YYYY/MM/DD")
	transCmd.AddCommand(transLsCmd)
	rootCmd.AddCommand(transCmd)
}

var dateRegex = regexp.MustCompile("[0-9]{4}/[0-9]{2}/[0-9]{2}")

func validateDateStr(dateStr string) bool {
	return dateRegex.Match([]byte(dateStr))
}

func lsTransactions(cmd *cobra.Command, args []string) {
	var err error
	var queryTerms []string
	var url string

	startDateStr, errStart := cmd.Flags().GetString("start")
	endDateStr, errEnd := cmd.Flags().GetString("end")
	if errEnd == nil && endDateStr != "" {
		if validateDateStr(endDateStr) {
			queryTerms = append(queryTerms, fmt.Sprintf("endDate=%s", endDateStr))
		} else {
			fmt.Println("End specified but is invalid. Will ignore.")
			endDateStr = ""
		}
	}
	if errStart == nil && startDateStr != "" {
		if validateDateStr(startDateStr) {
			queryTerms = append(queryTerms, fmt.Sprintf("startDate=%s", startDateStr))
		} else {
			fmt.Println("Start specified but is invalid. Will ignore.")
			startDateStr = ""
		}
	}
	switch {
	case startDateStr == "" && endDateStr != "":
		fmt.Println("Only end is specified! Will discard.")
		queryTerms = nil
	case startDateStr != "" && endDateStr == "":
		fmt.Println("Only start is specified! Will use today as the end!")
		queryTerms = append(
			queryTerms,
			fmt.Sprintf("endDate=%s", time.Now().Format("2006/01/02")),
		)
	}

	url = BASE_URL + "transactions"
	if len(queryTerms) > 0 {
		url += "?" + strings.Join(queryTerms, "&")
	}

	var transactions []bookkeeper.Transaction_
	resp, err := http.Get(url)
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
