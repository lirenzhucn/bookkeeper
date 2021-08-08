package cmd

import (
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

type JournalEntry struct {
	bookkeeper.JournalEntry
}

type CategoryMap []struct {
	Category      string   `json:"category"`
	SubCategories []string `json:"sub_categories"`
}

func (entry *JournalEntry) InteractiveSingleExpenseIncome(
	accounts []bookkeeper.Account, categoryMap CategoryMap,
) (err error) {
	var accountNames []string
	for _, a := range accounts {
		accountNames = append(accountNames, a.Name)
	}
	var categories []string
	var allSubCategories []string
	for _, c := range categoryMap {
		categories = append(categories, c.Category)
		allSubCategories = append(allSubCategories, c.SubCategories...)
	}
	entry.Clear()
	var qs []*survey.Question
	qs = []*survey.Question{
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "A quick title of the journal entry?",
				Default: "Single Expense / Income",
			},
		},
		{
			Name:   "desc",
			Prompt: &survey.Input{Message: "A more detailed description"},
		},
		{
			Name: "date",
			Prompt: &survey.Input{
				Message: "When did this transaction happen?",
				Default: time.Now().UTC().Format("2006/01/02"),
			},
			Validate: func(ans interface{}) error {
				str, _ := ans.(string)
				_, err := time.Parse("2006/01/02", str)
				return err
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				str, _ := ans.(string)
				newAns, _ = time.Parse("2006/01/02", str)
				return
			},
		},
		{
			Name: "type",
			Prompt: &survey.Select{
				Message: "Choose a transaction type:",
				Options: bookkeeper.VALID_TRANSACTION_TYPES,
				Default: "Out",
			},
		},
		{
			Name: "account_name",
			Prompt: &survey.Select{
				Message: "What is the account used?",
				Options: accountNames,
				Default: accountNames[0],
			},
		},
		{
			Name: "category",
			Prompt: &survey.Select{
				Message: "What is the category of the transaction?",
				Options: categories,
				Default: categories[0],
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				c, _ := ans.(survey.OptionAnswer)
				for _, q := range qs {
					if q.Name == "sub_category" {
						p, ok := q.Prompt.(*survey.Select)
						if !ok {
							continue
						}
						p.Options = categoryMap[c.Index].SubCategories
						p.Default = categoryMap[c.Index].SubCategories[0]
					}
				}
				newAns = ans
				return
			},
		},
		{
			Name: "sub_category",
			Prompt: &survey.Select{
				Message: "What is the sub-category of the transaction?",
				Options: allSubCategories,
				Default: allSubCategories[0],
			},
			Validate: survey.Required,
		},
		{
			Name: "Amount",
			Prompt: &survey.Input{
				Message: "What is the amount?",
				Default: "0.0",
			},
			Validate: func(ans interface{}) error {
				str, _ := ans.(string)
				_, err := strconv.ParseFloat(str, 64)
				return err
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				str, _ := ans.(string)
				val, _ := strconv.ParseFloat(str, 64)
				newAns = int64(val * 100)
				return
			},
		},
	}
	answers := struct {
		Title       string
		Desc        string
		Date        time.Time
		Type        string
		AccountName string `survey:"account_name"`
		Category    string
		SubCategory string `survey:"sub_category"`
		Amount      int64
	}{}
	if err = survey.Ask(qs, &answers); err != nil {
		return
	}
	entry.Transactions = append(entry.Transactions, bookkeeper.Transaction_{
		Transaction: bookkeeper.Transaction{
			Type:        answers.Type,
			Date:        answers.Date,
			Category:    answers.Category,
			SubCategory: answers.SubCategory,
			Amount:      answers.Amount,
			Notes:       answers.Title + ";" + answers.Desc,
		},
		AccountName: answers.AccountName,
	})
	entry.VerifyAndFillAccountIds(accounts)
	return
}

func (entry *JournalEntry) InteractiveConfirm() bool {
	tablePrintTransactions(entry.Transactions)
	confirmed := false
	survey.AskOne(&survey.Confirm{
		Message: "Are you sure to post these transactions?",
		Default: confirmed,
	}, &confirmed)
	return confirmed
}

func (entry *JournalEntry) PostToServer() error {
	for _, trans := range entry.Transactions {
		_, err := postSingleTransaction(trans.Transaction)
		if err != nil {
			return err
		}
	}
	return nil
}
