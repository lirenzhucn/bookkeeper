package bookkeeper

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func InitDb(dbpool *pgxpool.Pool, dryRun bool) ([]string, error) {
	var (
		tx       pgx.Tx
		err      error
		commands []string
	)
	commands = append(commands, "drop table if exists transactions;")
	commands = append(commands, "drop table if exists accounts;")
	commands = append(commands, GetSqlCreateAccounts())
	commands = append(commands, GetSqlCreateTransactions())
	if !dryRun {
		log.Println("Executing the following commands")
		for _, c := range commands {
			log.Println(c)
		}
		tx, err = dbpool.Begin(context.Background())
		if err != nil {
			return commands, err
		}
		defer tx.Rollback(context.Background())
		err = execCommands(tx, commands)
	}
	return commands, err
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
