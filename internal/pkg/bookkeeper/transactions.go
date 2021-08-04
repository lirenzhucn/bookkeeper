package bookkeeper

import (
	"strings"
	"time"
)

type Transaction struct {
	Id            int       `json:"id"`   // Unique id of the transaction
	Type          string    `json:"type"` // Transaction type, see validation for allowed values
	Date          time.Time `json:"date"`
	Category      string    `json:"category"`     // Tier 1 of the 2-tiered category
	SubCategory   string    `json:"sub_category"` // Tier 2 of the 2-tiered category
	AccountId     int       `json:"account_id"`
	Amount        int64     `json:"amount"`
	Notes         string    `json:"notes"`
	AssociationId string    `json:"association_id"` // Links TransferIn with TransferOut
}

type Transaction_ struct {
	Transaction
	AccountName string `json:"account_name"`
}

var VALID_TRANSACTION_TYPES = []string{
	"TransferIn", "TransferOut", "In", "Out", "BalanceChange", "LiabilityChange",
}

func (trans Transaction) Validate() bool {
	valid := stringInList(trans.Type, VALID_TRANSACTION_TYPES)
	switch {
	case strings.HasPrefix(trans.Type, "Transfer"):
		valid = valid && trans.AssociationId != ""
	case trans.Type == "In" || trans.Type == "Out":
		valid = valid && trans.Category != "" && trans.SubCategory != ""
	}
	return valid
}

func GetSqlCreateTransactions() string {
	return `create table transactions (
		id             serial,
		type           text,
		date           timestamp,
		category       text,
		sub_category   text,
		account_id     int,
		amount         bigint,
		notes          text,
		association_id text,
		primary key(id),
		constraint fk_account
			foreign key(account_id)
				references accounts(id)
	);`
}
