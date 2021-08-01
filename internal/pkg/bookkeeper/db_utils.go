package bookkeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
	}
	return commands, err
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
