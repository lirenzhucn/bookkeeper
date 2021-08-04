package bookkeeper

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func ComputeAccountBalance(
	dbpool *pgxpool.Pool, accountName string, date time.Time,
) (Account, int64, error) {
	var (
		account Account
		amount  int64
		err     error
		row     pgx.Row
	)
	row = dbpool.QueryRow(
		context.Background(),
		"select id, name, desc_, tags from accounts where name = $1 limit 1",
		accountName,
	)
	err = row.Scan(&account.Id, &account.Name, &account.Desc, &account.Tags)
	if err != nil {
		return account, amount, err
	}
	row = dbpool.QueryRow(
		context.Background(),
		`select sum(t.amount) from transactions t
inner join accounts a on t.account_id = a.id
where a.name = $1 and t.date <= $2`,
		accountName, date,
	)
	err = row.Scan(&amount)
	if err != nil {
		return account, amount, err
	}
	return account, amount, err
}
