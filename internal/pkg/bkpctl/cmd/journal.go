package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	SalaryIncomeJournal
)

var JournalTypeFlagIds = map[JournalTypeFlag][]string{
	SingleExpenseIncomeJournal: {"single"},
	SalaryIncomeJournal:        {"salary"},
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

func getAllAcounts(accounts *[]bookkeeper.Account) error {
	url_ := BASE_URL + "accounts"
	resp, err := http.Get(url_)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get accounts; response status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, accounts)
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
	getAllAcounts(&accounts)
	var entry JournalEntry
	switch journalTypeFlag {
	case SingleExpenseIncomeJournal:
		err := entry.InteractiveSingleExpenseIncome(accounts, categoryMap)
		cobra.CheckErr(err)
	case SalaryIncomeJournal:
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
