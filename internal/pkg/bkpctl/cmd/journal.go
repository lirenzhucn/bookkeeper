package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Record a journal entry",
	Run:   recordActivity,
}

type JournalTypeFlag enumflag.Flag

const (
	SingleExpenseIncomeJournal JournalTypeFlag = iota
	PaycheckJournal
	SingleTransferJournal
	InvestActivityJournal
)

var JournalTypeFlagIds = map[JournalTypeFlag][]string{
	SingleExpenseIncomeJournal: {"single"},
	PaycheckJournal:            {"paycheck"},
	SingleTransferJournal:      {"transfer"},
	InvestActivityJournal:      {"invest", "investment", "gain", "loss"},
}

var journalTypeFlag JournalTypeFlag

func initRecordCmd(rootCmd *cobra.Command) {
	journalCmd.Flags().VarP(
		enumflag.New(
			&journalTypeFlag, "type", JournalTypeFlagIds,
			enumflag.EnumCaseInsensitive,
		),
		"type", "t", "Type of the journal entry",
	)
	journalCmd.MarkFlagRequired("type")
	journalCmd.Flags().StringP(
		"categories", "c", "",
		"Path to the Category definition file (default: ./configs/category_map.json)",
	)
	journalCmd.Flags().StringP(
		"payroll-config", "p", "",
		"Path to the Payroll config file (default: ./configs/tpl/ws_payroll.json)",
	)
	rootCmd.AddCommand(journalCmd)
}

func readCategoryMap(categoriesFile string, categoryMap *CategoryMap) error {
	if categoriesFile == "" {
		categoriesFile = "configs/category_map.json"
	}
	f, err := os.Open(categoriesFile)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(categoryMap)
	if err != nil {
		return err
	}
	return nil
}

func recordActivity(cmd *cobra.Command, args []string) {
	// read category map
	categoriesFile, err := cmd.Flags().GetString("categories")
	cobra.CheckErr(err)
	var categoryMap CategoryMap
	readCategoryMap(categoriesFile, &categoryMap)
	// get all accounts
	var accounts []bookkeeper.Account
	getAllAccounts(&accounts)
	var entry JournalEntry
	switch journalTypeFlag {
	case SingleExpenseIncomeJournal:
		err := entry.InteractiveSingleExpenseIncome(accounts, categoryMap)
		cobra.CheckErr(err)
	case PaycheckJournal:
		err := entry.InteractivePaycheck(accounts, categoryMap)
		cobra.CheckErr(err)
	case SingleTransferJournal:
		err := entry.InteractiveTransfer(accounts)
		cobra.CheckErr(err)
	case InvestActivityJournal:
		err := entry.InteractiveInvest(accounts, categoryMap,
			func(accountName string, date string) (balance int64, err error) {
				account_, err := singleAccountBalance(accountName, date)
				balance = account_.Balance
				return
			})
		cobra.CheckErr(err)
	default:
		cobra.CheckErr(fmt.Errorf("invalid journal type %d", journalTypeFlag))
	}
	if entry.InteractiveConfirm() {
		fmt.Println("Posting the journal entry to the server...")
		err = entry.PostToServer()
		cobra.CheckErr(err)
	} else {
		fmt.Println("No journal entries or transactions are posted.")
	}
}
