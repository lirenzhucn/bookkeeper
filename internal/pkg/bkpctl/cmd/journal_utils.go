package cmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/leekchan/accounting"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

type JournalEntry struct {
	bookkeeper.JournalEntry
}

type CategoryMap []struct {
	Category      string   `json:"category"`
	SubCategories []string `json:"sub_categories"`
}

func (cm CategoryMap) GetAllCategories() []string {
	var categories []string
	for _, c := range cm {
		categories = append(categories, c.Category)
	}
	return categories
}

func (cm CategoryMap) GetAllSubCategories() []string {
	var allSubCategories []string
	for _, c := range cm {
		allSubCategories = append(allSubCategories, c.SubCategories...)
	}
	return allSubCategories
}

func (cm CategoryMap) GetAllSubCategoriesFullNames() []string {
	var allSubCategories []string
	for _, c := range cm {
		for _, sc := range c.SubCategories {
			allSubCategories = append(allSubCategories, c.Category+"/"+sc)
		}
	}
	return allSubCategories
}

func (cm CategoryMap) GetSubCategoriesByIndex(ind int) []string {
	return cm[ind].SubCategories
}

func (cm CategoryMap) GetSubCategoriesByName(category string) []string {
	for _, c := range cm {
		if c.Category == category {
			return c.SubCategories
		}
	}
	return nil
}

func getTodayNoTimeZone() time.Time {
	today, _ := time.Parse("2006/01/02", time.Now().Format("2006/01/02"))
	return today
}

type TransferBasicAnswerType struct {
	Title           string
	Desc            string
	Date            time.Time
	FromAccountName string
	ToAccountName   string
	Amount          int64
}

func (entry *JournalEntry) interactiveTransferEntryBasic(
	accountNames []string,
	answers *TransferBasicAnswerType,
	skipIfPreset bool,
) (err error) {
	var qs []*survey.Question
	defaultDate := getTodayNoTimeZone()
	if !reflect.ValueOf(answers.Date).IsZero() {
		defaultDate = answers.Date
	}
	if !skipIfPreset || reflect.ValueOf(answers.Title).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "title",
			Prompt: &survey.Input{
				Message: "A quick title of this transfer?",
				Default: answers.Title,
			},
		})
	}
	if !skipIfPreset || reflect.ValueOf(answers.Desc).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "desc",
			Prompt: &survey.Input{
				Message: "A more detailed description",
				Default: answers.Desc,
			},
		})
	}
	if !skipIfPreset || reflect.ValueOf(answers.Date).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "date",
			Prompt: &survey.Input{
				Message: "When did this transfer happen?",
				Default: defaultDate.Format("2006/01/02"),
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
		})
	}
	if !skipIfPreset || reflect.ValueOf(answers.FromAccountName).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "fromaccountname",
			Prompt: &survey.Select{
				Message: "From which account is this tranfer initiated?",
				Options: accountNames,
				Default: answers.FromAccountName,
			},
		})
	}
	if !skipIfPreset || reflect.ValueOf(answers.ToAccountName).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "toaccountname",
			Prompt: &survey.Select{
				Message: "To which account is this tranfer destinated?",
				Options: accountNames,
				Default: answers.ToAccountName,
			},
		})
	}
	// always ask for amount
	qs = append(qs, &survey.Question{
		Name: "amount",
		Prompt: &survey.Input{
			Message: "What is the amount?",
			Default: fmt.Sprintf("%.2f", float64(answers.Amount)/100),
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
	})
	if err = survey.Ask(qs, answers); err != nil {
		return
	}
	u, err := uuid.NewUUID()
	if err != nil {
		return
	}
	associationId := u.String()
	entry.Transactions = append(entry.Transactions, bookkeeper.Transaction_{
		Transaction: bookkeeper.Transaction{
			Type:          "TransferOut",
			Date:          answers.Date,
			Amount:        -answers.Amount, // flip the amount for the FROM account
			Notes:         answers.Title + ";" + answers.Desc + ";From",
			AssociationId: associationId,
		},
		AccountName: answers.FromAccountName,
	})
	entry.Transactions = append(entry.Transactions, bookkeeper.Transaction_{
		Transaction: bookkeeper.Transaction{
			Type:          "TransferIn",
			Date:          answers.Date,
			Amount:        answers.Amount,
			Notes:         answers.Title + ";" + answers.Desc + ";To",
			AssociationId: associationId,
		},
		AccountName: answers.ToAccountName,
	})
	return
}

type TransactionBasicAnswerType struct {
	Title       string
	Desc        string
	Date        time.Time
	Type        string
	AccountName string
	Category    string
	SubCategory string
	Amount      int64
}

func (entry *JournalEntry) interactiveJournalEntryBasic(
	accountNames []string, categoryMap CategoryMap,
	answers *TransactionBasicAnswerType,
) (err error) {
	var qs []*survey.Question
	defaultDate := getTodayNoTimeZone()
	if !reflect.ValueOf(answers.Date).IsZero() {
		defaultDate = answers.Date
	}
	qs = []*survey.Question{
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "A quick title of the journal entry?",
				Default: answers.Title,
			},
		},
		{
			Name: "desc",
			Prompt: &survey.Input{
				Message: "A more detailed description",
				Default: answers.Desc,
			},
		},
		{
			Name: "date",
			Prompt: &survey.Input{
				Message: "When did this transaction happen?",
				Default: defaultDate.Format("2006/01/02"),
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
				Default: answers.Type,
			},
		},
		{
			Name: "accountname",
			Prompt: &survey.Select{
				Message: "What is the account used?",
				Options: accountNames,
				Default: answers.AccountName,
			},
		},
		{
			Name: "category",
			Prompt: &survey.Select{
				Message: "What is the category of the transaction?",
				Options: categoryMap.GetAllCategories(),
				Default: answers.Category,
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				// a hacky hook to narrow down sub-category
				c, _ := ans.(survey.OptionAnswer)
				for _, q := range qs {
					if q.Name == "subcategory" {
						p, ok := q.Prompt.(*survey.Select)
						if !ok {
							continue
						}
						p.Options = categoryMap.GetSubCategoriesByIndex(c.Index)
					}
				}
				newAns = ans
				return
			},
		},
		{
			Name: "subcategory",
			Prompt: &survey.Select{
				Message: "What is the sub-category of the transaction?",
				Options: categoryMap.GetAllSubCategories(),
				Default: answers.SubCategory,
			},
			Validate: survey.Required,
		},
		{
			Name: "amount",
			Prompt: &survey.Input{
				Message: "What is the amount?",
				Default: fmt.Sprintf("%.2f", float64(answers.Amount)/100),
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
	if err = survey.Ask(qs, answers); err != nil {
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
	return
}

// if a field in trans is not a zero value, the field will be skipped
func interactiveTransactionWithPresets(
	accountNames []string,
	categoryMap CategoryMap,
	trans *bookkeeper.Transaction_,
) (err error) {
	var qs []*survey.Question
	if reflect.ValueOf(trans.Type).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "type",
			Prompt: &survey.Select{
				Message: "Choose a transaction type",
				Options: bookkeeper.VALID_TRANSACTION_TYPES,
			},
		})
	}
	if reflect.ValueOf(trans.AccountName).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "accountname",
			Prompt: &survey.Select{
				Message: "What is the account used?",
				Options: accountNames,
			},
		})
	}
	if reflect.ValueOf(trans.Category).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "category",
			Prompt: &survey.Select{
				Message: "What is the category?",
				Options: categoryMap.GetAllCategories(),
			},
			Transform: func(ans interface{}) (newAns interface{}) {
				// a hacky hook to narrow down sub-category
				c, _ := ans.(survey.OptionAnswer)
				for _, q := range qs {
					if q.Name == "subcategory" {
						p, ok := q.Prompt.(*survey.Select)
						if !ok {
							continue
						}
						p.Options = categoryMap.GetSubCategoriesByIndex(c.Index)
					}
				}
				newAns = ans
				return
			},
		})
	}
	subCategories := categoryMap.GetAllSubCategories()
	if !reflect.ValueOf(trans.Category).IsZero() {
		for _, c := range categoryMap {
			if c.Category == trans.Category {
				subCategories = c.SubCategories
				break
			}
		}
	}
	if reflect.ValueOf(trans.SubCategory).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "subcategory",
			Prompt: &survey.Select{
				Message: "What is the sub-category",
				Options: subCategories,
			},
		})
	}
	// always ask about the amount
	qs = append(qs, &survey.Question{
		Name: "amount",
		Prompt: &survey.Input{
			Message: "What is the amount?",
			Default: "0.00",
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
	})
	if reflect.ValueOf(trans.Notes).IsZero() {
		qs = append(qs, &survey.Question{
			Name: "notes",
			Prompt: &survey.Input{
				Message: "Any additional notes?",
			},
		})
	}
	if err = survey.Ask(qs, trans); err != nil {
		return
	}
	// NOTE: flip any out-going transaction amounts
	if trans.Type == "TransferOut" || trans.Type == "Out" ||
		trans.Type == "LiabilityChange" {
		trans.Amount = -trans.Amount
	}
	return
}

func (entry *JournalEntry) interactivePaycheckTaxes(
	accountNames []string, categoryMap CategoryMap,
) (err error) {
	var trans bookkeeper.Transaction_
	trans.Type = "Out"
	if len(entry.Transactions) > 0 {
		trans.Date = entry.Transactions[0].Date
		trans.AccountName = entry.Transactions[0].AccountName
	}
	taxesCategoryInd := -1
	for i, c := range categoryMap {
		if strings.ToLower(c.Category) == "taxes" ||
			strings.ToLower(c.Category) == "tax" {
			taxesCategoryInd = i
		}
	}
	if taxesCategoryInd >= 0 {
		trans.Category = categoryMap[taxesCategoryInd].Category
	}
	err = interactiveTransactionWithPresets(accountNames, categoryMap, &trans)
	entry.Transactions = append(entry.Transactions, trans)
	if err != nil {
		return
	}
	return
}

func (entry *JournalEntry) interactivePaycheckMedicalInsurance(
	accountNames []string, categoryMap CategoryMap,
) (err error) {
	var trans bookkeeper.Transaction_
	trans.Type = "Out"
	if len(entry.Transactions) > 0 {
		// always assume the first transaction is the "primary" transaction
		trans.Date = entry.Transactions[0].Date
		trans.AccountName = entry.Transactions[0].AccountName
	}
	trans.Category = "Medical Exp"
	trans.SubCategory = "Health Insurance"
	err = interactiveTransactionWithPresets(accountNames, categoryMap, &trans)
	entry.Transactions = append(entry.Transactions, trans)
	if err != nil {
		return
	}
	return
}

func (entry *JournalEntry) balanceOnAccount(accountName string) int64 {
	var balance int64 = 0
	for _, trans := range entry.Transactions {
		if trans.AccountName == accountName {
			balance += trans.Amount
		}
	}
	return balance
}

func (entry *JournalEntry) interactivePaycheckOtherExp(
	accountNames []string, categoryMap CategoryMap,
) (err error) {
	var trans bookkeeper.Transaction_
	trans.Type = "Out"
	if len(entry.Transactions) > 0 {
		// always assume the first transaction is the "primary" transaction
		trans.Date = entry.Transactions[0].Date
		trans.AccountName = entry.Transactions[0].AccountName
	}
	trans.Category = "Other Exp"
	trans.SubCategory = "Misc Exp"
	err = interactiveTransactionWithPresets(accountNames, categoryMap, &trans)
	entry.Transactions = append(entry.Transactions, trans)
	return
}

func (entry *JournalEntry) InteractivePaycheck(
	accounts []bookkeeper.Account, categoryMap CategoryMap,
) (err error) {
	var accountNames []string
	for _, a := range accounts {
		accountNames = append(accountNames, a.Name)
	}
	entry.Clear()
	// add validators for this type of journal entry
	entry.Validators = []string{"transfer_match"}
	// some default values
	ansBasic := TransactionBasicAnswerType{
		Title:       "Paycheck",
		Type:        "In",
		Category:    "Professional Income",
		SubCategory: "Salary",
		AccountName: "WS PAYROLL",
	}
	colorHeading := color.New(color.FgCyan).Add(color.Bold).Add(color.Underline)
	colorHeading.Println("Some general info of the paycheck")
	err = entry.interactiveJournalEntryBasic(accountNames, categoryMap, &ansBasic)
	if err != nil {
		return
	}
	// add zero balance validtor for the primary account
	entry.Validators = append(entry.Validators, "zero_balance:"+ansBasic.AccountName)
	colorHeading.Println("Now, let's record some taxes...")
	// Taxes
	for {
		hasTaxes := false
		survey.AskOne(&survey.Confirm{
			Message: "Want to add one (more) tax expense?",
			Default: true,
		}, &hasTaxes)
		if !hasTaxes {
			break
		}
		if err = entry.interactivePaycheckTaxes(
			accountNames, categoryMap); err != nil {
			return
		}
	}
	colorHeading.Println("Let's record healthcare insurance expenses...")
	// Medical Exp
	for {
		hasMedical := false
		survey.AskOne(&survey.Confirm{
			Message: "Want to add one (more) medical insurance expense?",
			Default: true,
		}, &hasMedical)
		if !hasMedical {
			break
		}
		if err = entry.interactivePaycheckMedicalInsurance(
			accountNames, categoryMap); err != nil {
			return
		}
	}
	colorHeading.Println("Any other expenses that are on the paycheck?")
	// Other Exp
	for {
		hasOtherExp := false
		survey.AskOne(&survey.Confirm{
			Message: "Want to add one (more) other expense?",
			Default: true,
		}, &hasOtherExp)
		if !hasOtherExp {
			break
		}
		if err = entry.interactivePaycheckOtherExp(
			accountNames, categoryMap,
		); err != nil {
			return
		}
	}
	colorHeading.Println("Now, let's put our money to where they belong...")
	// Transfers
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	for {
		balance := entry.balanceOnAccount(ansBasic.AccountName)
		hasTransfer := false
		survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf(
				"Want to add one (more) transfer (%s remaining)?",
				ac.FormatMoney(float64(balance)/100),
			),
			Default: balance != 0,
		}, &hasTransfer)
		if !hasTransfer {
			break
		}
		ansTransfer := TransferBasicAnswerType{
			Date:            ansBasic.Date,
			FromAccountName: ansBasic.AccountName,
			Amount:          balance,
		}
		if err = entry.interactiveTransferEntryBasic(
			accountNames, &ansTransfer, true,
		); err != nil {
			return
		}
	}
	if err = entry.VerifyAndFillAccountIds(accounts); err != nil {
		return
	}
	if err = entry.Validate(); err != nil {
		return
	}
	return
}

func (entry *JournalEntry) InteractiveSingleExpenseIncome(
	accounts []bookkeeper.Account, categoryMap CategoryMap,
) (err error) {
	var accountNames []string
	for _, a := range accounts {
		accountNames = append(accountNames, a.Name)
	}
	entry.Clear()
	var answers TransactionBasicAnswerType
	answers.Title = "Single Expense / Income"
	answers.Type = "Out"
	err = entry.interactiveJournalEntryBasic(accountNames, categoryMap, &answers)
	if err != nil {
		return
	}
	// finally check account ids
	err = entry.VerifyAndFillAccountIds(accounts)
	if err != nil {
		return
	}
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
