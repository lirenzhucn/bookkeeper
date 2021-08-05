package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/leekchan/accounting"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate various financial reports",
}
var balanceSheetCmd = &cobra.Command{
	Use:   "balance-sheet",
	Short: "Generate the balance sheet of a specified date",
	Args:  cobra.NoArgs,
	Run:   generateBalanceSheet,
}

func initReportCmd(rootCmd *cobra.Command) {
	balanceSheetCmd.Flags().StringSliceP(
		"asset-tags", "a",
		[]string{"cash", "taxable+liquid", "retirement", "education",
			"nonliquid", "real estate"},
		"Specify a list of asset tags to report separately",
	)
	balanceSheetCmd.Flags().StringSliceP(
		"liability-tags", "l",
		[]string{"credit card", "loan"},
		"Specify a list of liability tags to report separately",
	)
	balanceSheetCmd.Flags().StringP(
		"report-schema", "r",
		`{
"mapping": {
	"Cash or Equivalent": ["Assets/cash"],
	"Taxable Securities": ["Assets/taxable+liquid"],
	"Liquid Assets": ["Assets/cash", "Assets/taxable+liquid"],
	"Retirement Savings": ["Assets/retirement"],
	"Education Savings": ["Assets/education"],
	"Other Nonliquid": ["Assets/nonliquid"],
	"Assets excl. Home": [
		"Assets/cash", "Assets/taxable+liquid", "Assets/retirement", "Assets/education", "Assets/nonliquid"
	],
	"Home Net Mortgage": ["Assets/real estate"],
	"Total Assets": ["Assets/TOTAL"],
	"Short Term Liabilities": ["Liabilities/credit card"],
	"Long Term Liabilities": ["Liabilities/loan"],
	"Total Liabilities": ["Liabilities/TOTAL"],
	"Stockholders Equities": ["Equities/TOTAL"]
},
"order": [
	"Cash or Equivalent",
	"Taxable Securities",
	"Liquid Assets",
	"Retirement Savings",
	"Education Savings",
	"Other Nonliquid",
	"Assets excl. Home",
	"Home Net Mortgage",
	"Total Assets",
	"Short Term Liabilities",
	"Long Term Liabilities",
	"Total Liabilities",
	"Stockholders Equities"
],
"formatters": {
	"Liquid Assets": ["bold"],
	"Assets excl. Home": ["bold"],
	"Total Assets": ["bold", "green"],
	"Total Liabilities": ["bold", "red"],
	"Stockholders Equities": ["bold"]
}
}`,
		"Specify a report schema using a JSON string",
	)
	balanceSheetCmd.Flags().StringP("date", "d", "",
		"Specify the date to create the balance sheet for (default: today)")
	reportCmd.AddCommand(balanceSheetCmd)
	rootCmd.AddCommand(reportCmd)
}

type ReportSchema struct {
	Mapping    map[string][]string `json:"mapping"`
	Order      []string            `json:"order"`
	Formatters map[string][]string `json:"formatters"`
}

func generateBalanceSheet(cmd *cobra.Command, args []string) {
	assetTags, err := cmd.Flags().GetStringSlice("asset-tags")
	cobra.CheckErr(err)
	liabilityTags, err := cmd.Flags().GetStringSlice("liability-tags")
	cobra.CheckErr(err)
	dateStr, _ := cmd.Flags().GetString("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006/01/02")
	}
	reportSchemaStr, err := cmd.Flags().GetString("report-schema")
	cobra.CheckErr(err)
	var reportSchema ReportSchema
	err = json.Unmarshal([]byte(reportSchemaStr), &reportSchema)
	cobra.CheckErr(err)

	url_ := fmt.Sprintf(
		"%sreporting/balance_sheet?date=%s&assetTags=%s&liabilityTags=%s",
		BASE_URL, url.QueryEscape(dateStr),
		url.QueryEscape(strings.Join(assetTags, ",")),
		url.QueryEscape(strings.Join(liabilityTags, ",")),
	)
	resp, err := http.Get(url_)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	var balanceSheet bookkeeper.BalanceSheet
	json.NewDecoder(resp.Body).Decode(&balanceSheet)
	err = printBalanceSheet(balanceSheet, reportSchema, dateStr)
	cobra.CheckErr(err)
}

func buildTablewriterColors(formatters []string) tablewriter.Colors {
	var colors tablewriter.Colors
	for _, f := range formatters {
		switch f {
		case "bold":
			colors = append(colors, tablewriter.Bold)
		case "green":
			colors = append(colors, tablewriter.FgGreenColor)
		case "red":
			colors = append(colors, tablewriter.FgRedColor)
		}
	}
	return colors
}

func printBalanceSheet(
	balanceSheet bookkeeper.BalanceSheet,
	reportSchema ReportSchema,
	dateStr string,
) error {
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"", dateStr})
	// append data
	for _, itemName := range reportSchema.Order {
		tags, ok := reportSchema.Mapping[itemName]
		if !ok {
			return fmt.Errorf("missing item (%s) in mapping", itemName)
		}
		row := []string{itemName}
		var itemValue int64 = 0
		for _, tag := range tags {
			parts := strings.Split(tag, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag (%s) in schema", tag)
			}
			switch parts[0] {
			case "Assets":
				if parts[1] == "TOTAL" {
					itemValue += balanceSheet.Assets.Total
				} else {
					// if parts[1] is not in Groups, it will add 0
					itemValue += balanceSheet.Assets.Groups[parts[1]]
				}
			case "Liabilities":
				if parts[1] == "TOTAL" {
					itemValue += balanceSheet.Liabilities.Total
				} else {
					// if parts[1] is not in Groups, it will add 0
					itemValue += balanceSheet.Liabilities.Groups[parts[1]]
				}
			case "Equities":
				if parts[1] == "TOTAL" {
					itemValue += balanceSheet.Equities
				}
			default:
				return fmt.Errorf("invalid tag (%s) in schema", tag)
			}
		}
		row = append(row, ac.FormatMoney(float64(itemValue)/100))
		itemFormatter, ok := reportSchema.Formatters[itemName]
		if ok {
			cellColors := buildTablewriterColors(itemFormatter)
			table.Rich(row, []tablewriter.Colors{cellColors, cellColors})
		} else {
			table.Append(row)
		}
	}
	table.Render()
	return nil
}
