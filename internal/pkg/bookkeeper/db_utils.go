package bookkeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func InitDb(dbpool *pgxpool.Pool, dataFile string, dryRun bool) ([]string, error) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()
	var (
		tx       pgx.Tx
		err      error
		commands []string
		dbDump   DbDump
	)
	// drop and create tables
	commands = append(commands, "drop table if exists transactions;")
	commands = append(commands, "drop table if exists accounts;")
	commands = append(commands, GetSqlCreateAccounts())
	commands = append(commands, GetSqlCreateTransactions())
	// read accounts and transactions data
	if dataFile != "" {
		err = loadDataFromFile(dataFile, &dbDump)
		if err != nil {
			return commands, err
		}
		if dryRun {
			// if dry run, we return some description of what are about to be inserted
			commands = append(commands,
				fmt.Sprintf("<To insert %d account(s)>", len(dbDump.Accounts)))
			commands = append(commands,
				fmt.Sprintf("<To insert %d transactions(s)>",
					len(dbDump.Transactions)))
		}
	}
	// execute commands
	if !dryRun {
		for _, c := range commands {
			sugar.Infow("Execute command", "command", c)
		}
		tx, err = dbpool.Begin(context.Background())
		if err != nil {
			return commands, err
		}
		defer tx.Rollback(context.Background())
		// execute drop and create table commands
		if err := execCommands(tx, commands); err != nil {
			return commands, err
		}
		// insert records
		if err := insertAccounts(tx, dbDump.Accounts); err != nil {
			return commands, err
		}
		if err := insertTransactions(tx, dbDump.Transactions); err != nil {
			return commands, err
		}
		// reset sequence counts
		_, err = tx.Exec(
			context.Background(),
			"select setval('accounts_id_seq', coalesce((select max(id)+1 from accounts), 1), false)",
		)
		if err != nil {
			return commands, err
		}
		_, err = tx.Exec(
			context.Background(),
			"select setval('transactions_id_seq', coalesce((select max(id)+1 from transactions), 1), false)",
		)
		if err != nil {
			return commands, err
		}
	}
	return commands, err
}

func GetTransactionsBetweenDates(
	dbpool *pgxpool.Pool, start time.Time, end time.Time, limit int,
) ([]Transaction_, error) {
	var (
		transactions []Transaction_
		curr         Transaction_
	)
	rows, err := dbpool.Query(
		context.Background(),
		`select t.id, type, date, category, sub_category, account_id, amount, notes, association_id, a.name
from transactions t
inner join accounts a on t.account_id = a.id
where date >= $1 and date < $2
order by date desc
limit $3`,
		start,
		end,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&curr.Id,
			&curr.Type,
			&curr.Date,
			&curr.Category,
			&curr.SubCategory,
			&curr.AccountId,
			&curr.Amount,
			&curr.Notes,
			&curr.AssociationId,
			&curr.AccountName,
		); err != nil {
			return transactions, err
		}
		transactions = append(transactions, curr)
	}
	return transactions, nil
}

func GetAllTransactions(dbpool *pgxpool.Pool, limit int, offset int) ([]Transaction_, error) {
	var (
		transactions []Transaction_
		curr         Transaction_
	)
	rows, err := dbpool.Query(
		context.Background(),
		`select t.id, type, date, category, sub_category, account_id, amount, notes, association_id, a.name
from transactions t
inner join accounts a on t.account_id = a.id
order by date desc
limit $1 offset $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&curr.Id,
			&curr.Type,
			&curr.Date,
			&curr.Category,
			&curr.SubCategory,
			&curr.AccountId,
			&curr.Amount,
			&curr.Notes,
			&curr.AssociationId,
			&curr.AccountName,
		); err != nil {
			return transactions, err
		}
		transactions = append(transactions, curr)
	}
	return transactions, nil
}

func GetAllAccounts(dbpool *pgxpool.Pool, limit int, offset int) ([]Account, error) {
	var (
		accounts []Account
		curr     Account
	)
	rows, err := dbpool.Query(context.Background(), "select id, name, desc_, tags from accounts limit $1 offset $2", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&curr.Id, &curr.Name, &curr.Desc, &curr.Tags); err != nil {
			return accounts, err
		}
		accounts = append(accounts, curr)
	}
	return accounts, nil
}

func GetSingleAccount(dbpool *pgxpool.Pool, id int) (Account, error) {
	var account Account
	row := dbpool.QueryRow(context.Background(), "select id, name, desc_, tags from accounts where id = $1", id)
	err := row.Scan(&account.Id, &account.Name, &account.Desc, &account.Tags)
	return account, err
}

func GetSingleTransaction(dbpool *pgxpool.Pool, id int) (Transaction_, error) {
	var transaction Transaction_
	row := dbpool.QueryRow(
		context.Background(),
		`select t.id, type, date, category, sub_category, account_id, amount, notes, association_id, a.name
from transaction t
inner join accounts a on t.account_id = a.id
where t.id = $1`,
		id,
	)
	err := row.Scan(
		&transaction.Id, &transaction.Type, &transaction.Date,
		&transaction.Category, &transaction.SubCategory, &transaction.AccountId,
		&transaction.Amount, &transaction.Notes, &transaction.AssociationId,
		&transaction.AccountName,
	)
	return transaction, err
}

func InsertAccount(dbpool *pgxpool.Pool, account *Account) error {
	row := dbpool.QueryRow(
		context.Background(),
		`insert into accounts (name, desc_, tags) values ($1, $2, $3)
returning id, name, desc_, tags`,
		account.Name, account.Desc, account.Tags,
	)
	err := row.Scan(&account.Id, &account.Name, &account.Desc, &account.Tags)
	return err
}

func UpdateAccount(dbpool *pgxpool.Pool, account *Account) error {
	row := dbpool.QueryRow(
		context.Background(),
		`update accounts set name = $1, desc_ = $2, tags = $3 where id = $4
returning id, name, desc_, tags`,
		account.Name, account.Desc, account.Tags, account.Id,
	)
	err := row.Scan(&account.Id, &account.Name, &account.Desc, &account.Tags)
	return err
}

func DeleteAccount(dbpool *pgxpool.Pool, account_id int) error {
	_, err := dbpool.Exec(
		context.Background(),
		"delete from accounts where id = $1",
		account_id,
	)
	return err
}

func InsertTransaction(dbpool *pgxpool.Pool, trans *Transaction) error {
	row := dbpool.QueryRow(
		context.Background(),
		`insert into transactions
(type, date, category, sub_category, account_id, amount, notes, association_id)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning id, type, date, category, sub_category, account_id, amount, notes, association_id`,
		trans.Type, trans.Date, trans.Category, trans.SubCategory,
		trans.AccountId, trans.Amount, trans.Notes, trans.AssociationId,
	)
	err := row.Scan(
		&trans.Id, &trans.Type, &trans.Date, &trans.Category,
		&trans.SubCategory, &trans.AccountId, &trans.Amount, &trans.Notes,
		&trans.AssociationId,
	)
	return err
}

func DeleteTransaction(dbpool *pgxpool.Pool, trans_id int) error {
	_, err := dbpool.Exec(
		context.Background(),
		"delete from transactions where id = $1",
		trans_id,
	)
	return err
}

type DbDump struct {
	Accounts     []Account     `json:"accounts"`
	Transactions []Transaction `json:"transactions"`
}

func loadDataFromFile(dataFile string, dbDump *DbDump) error {
	jsonFile, err := os.Open(dataFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	json.Unmarshal(byteValue, dbDump)

	return nil
}

func insertAccounts(tx pgx.Tx, accounts []Account) error {
	for _, account := range accounts {
		_, err := tx.Exec(
			context.Background(),
			"insert into accounts (id, name, desc_, tags) values ($1, $2, $3, $4)",
			account.Id,
			account.Name,
			account.Desc,
			account.Tags,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertTransactions(tx pgx.Tx, transactions []Transaction) error {
	for _, transaction := range transactions {
		_, err := tx.Exec(
			context.Background(),
			`insert into transactions
(id, type, date, category, sub_category, account_id, amount, notes, association_id)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			transaction.Id,
			transaction.Type,
			transaction.Date,
			transaction.Category,
			transaction.SubCategory,
			transaction.AccountId,
			transaction.Amount,
			transaction.Notes,
			transaction.AssociationId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func execCommands(tx pgx.Tx, commands []string) error {
	for _, c := range commands {
		_, err := tx.Exec(context.Background(), c)
		if err != nil {
			return err
		}
	}
	tx.Commit(context.Background())
	return nil
}
