package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
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

var numDaysMatcher = regexp.MustCompile(`^past\s*(\d+)\s*day(s*)$`)

func parseQueryString(original string) (parsed string) {
	phrases := strings.SplitN(original, "on", 2)
	if len(phrases) < 2 {
		phrases = append(phrases, "")
	}
	dateRange, accountName := phrases[0], phrases[1]
	dateRange = strings.TrimSpace(dateRange)
	accountName = strings.TrimSpace(accountName)
	matchedGroups := numDaysMatcher.FindStringSubmatch(dateRange)
	switch {
	case dateRange == "past week":
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * 6)
		parsed = parsed + fmt.Sprintf(
			"date>=%s AND date<=%s", cutoff.Format("2006/01/02"),
			today.Format("2006/01/02"),
		)
	case dateRange == "last week":
		today := time.Now().Add(-time.Hour * 24 * 7)
		cutoff := today.Add(-time.Hour * 24 * 6)
		parsed = parsed + fmt.Sprintf(
			"date>=%s AND date<=%s", cutoff.Format("2006/01/02"),
			today.Format("2006/01/02"),
		)
	case matchedGroups != nil:
		numDays, _ := strconv.ParseInt(matchedGroups[1], 10, 64)
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * time.Duration(numDays-1))
		parsed = parsed + fmt.Sprintf(
			"date>=%s AND date<=%s", cutoff.Format("2006/01/02"),
			today.Format("2006/01/02"),
		)
	default:
		parsed = original
		return
	}
	if accountName != "" {
		parsed = parsed + fmt.Sprintf(` AND a.name="%s"`, accountName)
	}
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
