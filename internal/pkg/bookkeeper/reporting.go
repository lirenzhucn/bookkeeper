package bookkeeper

import (
	"context"
	"strings"
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
	if err != nil && strings.Contains(err.Error(), "cannot assign NULL") {
		// no transactions for this account
		return account, 0, nil
	}
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

type BalanceSheetItem struct {
	Total  int64            `json:"total"`
	Groups map[string]int64 `json:"groups"`
}

type BalanceSheet struct {
	Assets      BalanceSheetItem `json:"assets"`
	Liabilities BalanceSheetItem `json:"liabilities"`
	Equities    int64            `json:"equities"`
}

func ComputeBalanceSheet(
	accounts []AccountWithBalance, assetTags []string, liabilityTags []string,
) BalanceSheet {
	var balanceSheet BalanceSheet
	balanceSheet.Assets.Groups = make(map[string]int64)
	balanceSheet.Liabilities.Groups = make(map[string]int64)
	for _, account := range accounts {
		if stringInList("asset", account.Tags) ||
			stringInList("assets", account.Tags) {
			balanceSheet.Assets.Total += account.Balance
			for _, tag := range assetTags {
				if stringInList(tag, account.Tags) {
					oldAmount := balanceSheet.Assets.Groups[tag]
					balanceSheet.Assets.Groups[tag] = oldAmount + account.Balance
				}
			}
		}
		if stringInList("liability", account.Tags) ||
			stringInList("liabilities", account.Tags) {
			balanceSheet.Liabilities.Total -= account.Balance
			for _, tag := range liabilityTags {
				if stringInList(tag, account.Tags) {
					oldAmount := balanceSheet.Liabilities.Groups[tag]
					balanceSheet.Liabilities.Groups[tag] = oldAmount - account.Balance
				}
			}
		}
		balanceSheet.Equities = balanceSheet.Assets.Total - balanceSheet.Liabilities.Total
	}
	return balanceSheet
}
