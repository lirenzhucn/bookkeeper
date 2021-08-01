package bookkeeper

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func InitDb(dbpool *pgxpool.Pool, dryRun bool) ([]string, error) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()
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
		for _, c := range commands {
			sugar.Infow("Execute command", "command", c)
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
