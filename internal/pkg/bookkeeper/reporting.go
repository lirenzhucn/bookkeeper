package bookkeeper

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type AccountWithBalance struct {
	Account
	Balance int64 `json:"Balance"`
}

func ComputeAccountBalanceByName(
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

func ComputeAccountBalanceById(
	dbpool *pgxpool.Pool, accountId int, date time.Time,
) (Account, int64, error) {
	var (
		account Account
		amount  int64
		err     error
		row     pgx.Row
	)
	row = dbpool.QueryRow(
		context.Background(),
		"select id, name, desc_, tags from accounts where id = $1 limit 1",
		accountId,
	)
	err = row.Scan(&account.Id, &account.Name, &account.Desc, &account.Tags)
	if err != nil {
		return account, amount, err
	}
	row = dbpool.QueryRow(
		context.Background(),
		"select sum(amount) from transactions where account_id = $1 and date <= $2",
		accountId, date,
	)
	err = row.Scan(&amount)
	if err != nil {
		return account, amount, err
	}
	return account, amount, err
}

func GetAllAccountIds(dbpool *pgxpool.Pool) ([]int, error) {
	var ids []int
	rows, err := dbpool.Query(context.Background(), "select id from accounts")
	if err != nil {
		return ids, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
