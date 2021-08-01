package bookkeeper

import "time"

type Transaction struct {
	Id            int       `json:"Id"`   // Unique id of the transaction
	Type          string    `json:"Type"` // Transaction type, see validation for allowed values
	Date          time.Time `json:"Date"`
	Category      string    `json:"Category"`    // Tier 1 of the 2-tiered category
	SubCategory   string    `json:"SubCategory"` // Tier 2 of the 2-tiered category
	AccountId     int       `json:"AccountId"`
	Amount        int64     `json:"Amount"`
	Notes         string    `json:"Notes"`
	AssociationId string    `json:"AssociationId"` // Links TransferIn with TransferOut
}

var VALID_TRANSACTION_TYPES = []string{"TransferIn", "TransferOut", "In", "Out"}

func (trans Transaction) Validate() bool {
	return stringInList(trans.Type, VALID_TRANSACTION_TYPES)
}

func GetSqlCreateTransactions() string {
	return `create table transactions (
		id             serial,
		type           text,
		date           time,
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
