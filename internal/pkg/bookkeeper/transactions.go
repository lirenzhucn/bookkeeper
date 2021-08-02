package bookkeeper

import "time"

type Transaction struct {
	Id            int       `json:"id" header:"Id"`     // Unique id of the transaction
	Type          string    `json:"type" header:"Type"` // Transaction type, see validation for allowed values
	Date          time.Time `json:"date" header:"Date"`
	Category      string    `json:"category" header:"Category"`        // Tier 1 of the 2-tiered category
	SubCategory   string    `json:"sub_category" header:"SubCategory"` // Tier 2 of the 2-tiered category
	AccountId     int       `json:"account_id" header:"AccountId"`
	Amount        int64     `json:"amount" header:"Amount"`
	Notes         string    `json:"notes" header:"Notes"`
	AssociationId string    `json:"association_id" header:"AssociationId"` // Links TransferIn with TransferOut
}

var VALID_TRANSACTION_TYPES = []string{"TransferIn", "TransferOut", "In", "Out"}

func (trans Transaction) Validate() bool {
	return stringInList(trans.Type, VALID_TRANSACTION_TYPES)
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
