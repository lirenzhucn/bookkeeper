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
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Generate the balance sheet of a specified date",
	Args:  cobra.NoArgs,
	Run:   generateBalanceSheet,
}
var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "Generate the income statement of periods of time",
	Args:  cobra.NoArgs,
	Run:   generateIncomeStatement,
}

func initReportCmd(rootCmd *cobra.Command) {
	balanceCmd.Flags().StringSliceP(
		"asset-tags", "a",
		[]string{"cash", "taxable+liquid", "retirement", "education",
			"nonliquid", "real estate"},
		"Specify a list of asset tags to report separately",
	)
	balanceCmd.Flags().StringSliceP(
		"liability-tags", "l",
		[]string{"credit card", "loan"},
		"Specify a list of liability tags to report separately",
	)
	balanceCmd.Flags().StringP(
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
	balanceCmd.Flags().StringP("date", "d", "",
		"Specify the date to create the balance sheet for (default: today)")
	incomeCmd.Flags().StringP("date-range", "d", "",
		"Specify the date range to create the income statement for")
	incomeCmd.MarkFlagRequired("date-range")
	incomeCmd.Flags().StringSliceP(
		"revenue-tags", "r",
		[]string{
			"Professional Income/Salary", "Professional Income/RSU",
			"Professional Income/Employer Match", "Other Income/",
		},
		"Sepcify matchers for Revenue",
	)
	incomeCmd.Flags().StringSliceP(
		"taxes-tags", "t", []string{"Taxes/"}, "Sepcify matchers for Taxes",
	)
	incomeCmd.Flags().StringSliceP(
		"expenses-tags", "e",
		[]string{
			"Home/Mortgage Interest", "Home/Loan Fees", "Home/HOA",
			"Food & Dining/", "Kids/", "Bills & Utilities/", "Transportation/",
			"Entertainment/", "Shopping/", "Communications/", "Medical Exp/",
			"Other Exp/",
		},
		"Specify matchers for Expenses",
	)
	incomeCmd.Flags().StringSliceP(
		"investments-tags", "i",
		[]string{
			"Investment/Taxable Investment", "Investment/Retirement Investment",
			"Investment/Education Investment",
		},
		"Specify matchers for Investments",
	)
	incomeCmd.Flags().String(
		"report-schema",
		`{
"mapping": {
	"Salaries": ["Revenue/Professional Income/Salary"],
	"RSU": ["Revenue/Professional Income/RSU"],
	"Employer Match": ["Revenue/Professional Income/Employer Match"],
	"Other Revenue": ["Revenue/Other Income/"],
	"Total Revenue": ["Revenue/TOTAL"],
	"Tax Expenses": ["Taxes/TOTAL"],
	"Revenue Net Taxes": ["RevenueNetTaxes/TOTAL"],
	"Mortgage Interests": ["Expenses/Home/Mortgage Interest"],
	"Loan Fees": ["Expenses/Home/Loan Fees"],
	"HOA": ["Expenses/Home/HOA"],
	"Food & Dining": ["Expenses/Food & Dining/"],
	"Kids": ["Expenses/Kids/"],
	"Bills & Utilities": ["Expenses/Bills & Utilities/"],
	"Transportation": ["Expenses/Transportation/"],
	"Entertainment": ["Expenses/Entertainment/"],
	"Shopping": ["Expenses/Shopping/"],
	"Communications": ["Expenses/Communications/"],
	"Medical Exp": ["Expenses/Medical Exp/"],
	"Other Exp": ["Expenses/Other Exp/"],
	"Total Expenses": ["Expenses/TOTAL"],
	"Operating Income": ["OperatingIncome/TOTAL"],
	"Taxable Investments": ["Investments/Investment/Taxable Investment"],
	"Retirement Investments": ["Investments/Investment/Retirement Investment"],
	"Education Investments": ["Investments/Investment/Education Investment"],
	"Invest Gains/Losses": ["Investments/TOTAL"],
	"Total Earnings": ["TotalEarnings/TOTAL"]
},
"order": [
	"Salaries",
	"RSU",
	"Employer Match",
	"Other Revenue",
	"Total Revenue",
	"Tax Expenses",
	"Revenue Net Taxes",
	"Mortgage Interests",
	"Loan Fees",
	"HOA",
	"Food & Dining",
	"Kids",
	"Bills & Utilities",
	"Transportation",
	"Entertainment",
	"Shopping",
	"Communications",
	"Medical Exp",
	"Other Exp",
	"Total Expenses",
	"Operating Income",
	"Taxable Investments",
	"Retirement Investments",
	"Education Investments",
	"Invest Gains/Losses",
	"Total Earnings"
],
"formatters": {
	"Total Revenue": ["bold", "green"],
	"Revenue Net Taxes": ["bold", "green"],
	"Total Expenses": ["bold", "red"],
	"Operating Income": ["bold", "green"],
	"Invest Gains/Losses": ["bold", "green"],
	"Total Earnings": ["bold", "yellow"]
}
}`,
		"Specify a report schema using a JSON string",
	)
	reportCmd.AddCommand(balanceCmd)
	reportCmd.AddCommand(incomeCmd)
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
	var balanceSheets []bookkeeper.BalanceSheet
	json.NewDecoder(resp.Body).Decode(&balanceSheets)
	var statements []bookkeeper.StatementWithFields
	for _, bs := range balanceSheets {
		statements = append(statements, bs)
	}
	err = printStatements(statements, reportSchema, dateStr)
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
		case "yellow":
			colors = append(colors, tablewriter.FgYellowColor)
		}
	}
	return colors
}

func printStatements(
	statements []bookkeeper.StatementWithFields,
	reportSchema ReportSchema,
	dateStr string,
) error {
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{""}
	headers = append(headers, strings.Split(dateStr, ",")...)
	table.SetHeader(headers)
	// append data
	for _, itemName := range reportSchema.Order {
		tags, ok := reportSchema.Mapping[itemName]
		if !ok {
			return fmt.Errorf("missing item (%s) in mapping", itemName)
		}
		row := []string{itemName}
		for _, statement := range statements {
			var itemValue int64 = 0
			for _, tag := range tags {
				parts := strings.SplitN(tag, "/", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid tag (%s) in schema", tag)
				}
				rg, ok := statement.GetFieldAsReportGroup(parts[0])
				if ok {
					if parts[1] == "TOTAL" {
						itemValue += rg.Total
					} else {
						itemValue += rg.Groups[parts[1]]
					}
					continue
				}
				val, ok := statement.GetFieldAsInt64(parts[0])
				if ok {
					if parts[1] == "TOTAL" {
						itemValue += val
					}
					continue
				}
				return fmt.Errorf("invalid tag (%s) in schema", tag)
			}
			row = append(row, ac.FormatMoney(float64(itemValue)/100))
		}
		itemFormatter, ok := reportSchema.Formatters[itemName]
		if ok {
			cellColors := buildTablewriterColors(itemFormatter)
			var colorSlice []tablewriter.Colors
			for i := 0; i < len(row); i++ {
				colorSlice = append(colorSlice, cellColors)
			}
			table.Rich(row, colorSlice)
		} else {
			table.Append(row)
		}
	}
	table.Render()
	return nil
}

func generateIncomeStatement(cmd *cobra.Command, args []string) {
	dateRangeStr, err := cmd.Flags().GetString("date-range")
	cobra.CheckErr(err)
	revenueTags, err := cmd.Flags().GetStringSlice("revenue-tags")
	cobra.CheckErr(err)
	taxesTags, err := cmd.Flags().GetStringSlice("taxes-tags")
	cobra.CheckErr(err)
	expensesTags, err := cmd.Flags().GetStringSlice("expenses-tags")
	cobra.CheckErr(err)
	investmentsTags, err := cmd.Flags().GetStringSlice("investments-tags")
	cobra.CheckErr(err)
	reportSchemaStr, err := cmd.Flags().GetString("report-schema")
	cobra.CheckErr(err)
	var reportSchema ReportSchema
	err = json.Unmarshal([]byte(reportSchemaStr), &reportSchema)
	cobra.CheckErr(err)

	url_ := fmt.Sprintf(
		"%sreporting/income_statement?dateRange=%s&revenueTags=%s&taxesTags=%s&expensesTags=%s&investmentsTags=%s",
		BASE_URL, url.QueryEscape(dateRangeStr),
		url.QueryEscape(strings.Join(revenueTags, ",")),
		url.QueryEscape(strings.Join(taxesTags, ",")),
		url.QueryEscape(strings.Join(expensesTags, ",")),
		url.QueryEscape(strings.Join(investmentsTags, ",")),
	)
	resp, err := http.Get(url_)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	var isList []bookkeeper.IncomeStatement
	json.NewDecoder(resp.Body).Decode(&isList)
	var statements []bookkeeper.StatementWithFields
	for _, is := range isList {
		statements = append(statements, is)
	}
	err = printStatements(statements, reportSchema, dateRangeStr)
	cobra.CheckErr(err)
}
