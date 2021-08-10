package bookkeeper

import (
	"fmt"
	"strings"
	"time"
)

type JournalEntry struct {
	Title        string         `json:"title"`
	Desc         string         `json:"desc"`
	Transactions []Transaction_ `json:"transactions"`
	Validators   []string       `json:"validators"`
}

type JournalEntryValidationError struct {
	TransactionIdx int
	Validator      string
	Reason         string
	AssociationId  string
}

func (e JournalEntryValidationError) Error() (msg string) {
	switch {
	case e.TransactionIdx >= 0:
		msg = fmt.Sprintf(
			"validator %s failed on transaction %d with reason: %s",
			e.Validator, e.TransactionIdx, e.Reason,
		)
	case e.AssociationId != "":
		msg = fmt.Sprintf(
			"validator %s failed for association id %s with reason: %s",
			e.Validator, e.AssociationId, e.Reason,
		)
	default:
		msg = fmt.Sprintf(
			"validator %s failed with reason: %s", e.Validator, e.Reason,
		)
	}
	return
}

func (entry *JournalEntry) Init(numTransactions int) {
	entry.Validators = append(entry.Validators, "transfer_match")
	for i := 0; i < numTransactions; i++ {
		entry.Transactions = append(entry.Transactions,
			Transaction_{Transaction: Transaction{Date: time.Now().UTC()}})
	}
}

func (entry *JournalEntry) Validate() error {
	for _, validator := range entry.Validators {
		validator = strings.ToLower(validator)
		switch {
		case validator == "transfer_match":
			if err := entry.IsTransferMatch(); err != nil {
				return err
			}
		case strings.HasPrefix(validator, "zero_balance:"):
			accountName := strings.SplitN(validator, ":", 2)[1]
			if err := entry.IsZeroBalance(accountName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (entry *JournalEntry) IsTransferMatch() error {
	transferMap := make(map[string]int64)
	for i, trans := range entry.Transactions {
		if strings.HasPrefix(trans.Type, "Transfer") {
			if trans.AssociationId == "" {
				return JournalEntryValidationError{
					TransactionIdx: i,
					Validator:      "transfer_match",
					Reason:         "no association id",
					AssociationId:  "",
				}
			}
			oldVal := transferMap[trans.AssociationId]
			transferMap[trans.AssociationId] = oldVal + trans.Amount
		}
	}
	for i, val := range transferMap {
		if val != 0 {
			return JournalEntryValidationError{
				TransactionIdx: -1,
				Validator:      "transfer_match",
				Reason:         "unmatched transfer",
				AssociationId:  i,
			}
		}
	}
	return nil
}

func (entry *JournalEntry) IsZeroBalance(accountName string) error {
	var val int64 = 0
	for _, trans := range entry.Transactions {
		if trans.AccountName == accountName {
			val += trans.Amount
		}
	}
	if val != 0 {
		return JournalEntryValidationError{
			TransactionIdx: -1,
			Validator:      "zero_balance",
			Reason:         "non-zero balance for entry",
			AssociationId:  "",
		}
	}
	return nil
}

func (entry *JournalEntry) VerifyAndFillAccountIds(accounts []Account) error {
	accountIds := make(map[string]int)
	for _, account := range accounts {
		accountIds[account.Name] = account.Id
	}
	for i, trans := range entry.Transactions {
		accountId, present := accountIds[trans.AccountName]
		if !present {
			return fmt.Errorf(
				"account name %s is not found in valid accounts",
				trans.AccountName,
			)
		}
		entry.Transactions[i].AccountId = accountId
	}
	return nil
}

func (entry *JournalEntry) NumTransactions() int {
	return len(entry.Transactions)
}

func (entry *JournalEntry) Clear() {
	// set everything to 0 values first
	entry.Title = ""
	entry.Desc = ""
	entry.Transactions = nil
	entry.Validators = nil
}
