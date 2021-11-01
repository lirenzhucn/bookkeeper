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

type ReportGroup struct {
	Total  int64            `json:"total"`
	Groups map[string]int64 `json:"groups"`
}

func (rg *ReportGroup) Init() {
	rg.Groups = make(map[string]int64)
}

func (rg *ReportGroup) Remainder() int64 {
	res := rg.Total
	for _, v := range rg.Groups {
		res -= v
	}
	return res
}

type StatementWithFields interface {
	GetFieldAsReportGroup(fieldName string) (*ReportGroup, bool)
	GetFieldAsInt64(fieldName string) (int64, bool)
}

type BalanceSheet struct {
	Assets      ReportGroup `json:"assets"`
	Liabilities ReportGroup `json:"liabilities"`
	Equities    int64       `json:"equities"`
}

func (bs BalanceSheet) GetFieldAsReportGroup(fieldName string) (rg *ReportGroup, ok bool) {
	ok = true
	switch fieldName {
	case "Assets":
		rg = &bs.Assets
	case "Liabilities":
		rg = &bs.Liabilities
	default:
		ok = false
	}
	return
}

func (bs BalanceSheet) GetFieldAsInt64(fieldName string) (res int64, ok bool) {
	ok = true
	switch fieldName {
	case "Equities":
		res = bs.Equities
	default:
		ok = false
	}
	return
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
				match := tag != ""
				for _, t := range strings.Split(tag, "+") {
					match = match && stringInList(t, account.Tags)
				}
				if match {
					oldAmount := balanceSheet.Assets.Groups[tag]
					balanceSheet.Assets.Groups[tag] = oldAmount + account.Balance
				}
			}
		}
		if stringInList("liability", account.Tags) ||
			stringInList("liabilities", account.Tags) {
			balanceSheet.Liabilities.Total -= account.Balance
			for _, tag := range liabilityTags {
				match := tag != ""
				for _, t := range strings.Split(tag, "+") {
					match = match && stringInList(t, account.Tags)
				}
				if match {
					oldAmount := balanceSheet.Liabilities.Groups[tag]
					balanceSheet.Liabilities.Groups[tag] = oldAmount - account.Balance
				}
			}
		}
		balanceSheet.Equities = balanceSheet.Assets.Total - balanceSheet.Liabilities.Total
	}
	return balanceSheet
}

type IncomeStatement struct {
	Revenue         ReportGroup `json:"revenue"`
	Taxes           ReportGroup `json:"taxes"`
	RevenueNetTaxes int64       `json:"revenue_net_taxes"`
	Expenses        ReportGroup `json:"expenses"`
	OperatingIncome int64       `json:"operating_income"`
	Investments     ReportGroup `json:"investments"`
	TotalEarnings   int64       `json:"total_earnings"`
}

func (is IncomeStatement) GetFieldAsReportGroup(fieldName string) (rg *ReportGroup, ok bool) {
	ok = true
	switch fieldName {
	case "Revenue":
		rg = &is.Revenue
	case "Taxes":
		rg = &is.Taxes
	case "Expenses":
		rg = &is.Expenses
	case "Investments":
		rg = &is.Investments
	default:
		ok = false
	}
	return
}

func (is IncomeStatement) GetFieldAsInt64(fieldName string) (res int64, ok bool) {
	ok = true
	switch fieldName {
	case "RevenueNetTaxes":
		res = is.RevenueNetTaxes
	case "OperatingIncome":
		res = is.OperatingIncome
	case "TotalEarnings":
		res = is.TotalEarnings
	default:
		ok = false
	}
	return
}

func (is *IncomeStatement) Init() {
	is.Revenue.Init()
	is.Taxes.Init()
	is.Expenses.Init()
	is.Investments.Init()
}

func ComputeIncomeStatement(
	dbpool *pgxpool.Pool, startDate time.Time, endDate time.Time,
	revenueTags []string, taxesTags []string, expensesTags []string,
	investmentsTags []string,
) (is IncomeStatement, err error) {
	is.Init()
	rows, err := dbpool.Query(
		context.Background(),
		`select type, category, sub_category, amount from transactions
where date >= $1 and date <= $2`,
		startDate, endDate,
	)
	if err != nil {
		return
	}
	for rows.Next() {
		var (
			type_       string
			category    string
			subCategory string
			amount      int64
		)
		if err = rows.Scan(&type_, &category, &subCategory, &amount); err != nil {
			return
		}
		if type_ == "In" || type_ == "Out" {
			accumulateByTags(&is.Revenue, revenueTags, category+"/"+subCategory, amount)
			accumulateByTags(&is.Investments, investmentsTags, category+"/"+subCategory, amount)
			// NOTE: amounts for taxes and expenses are flipped, b/c they are
			// negative in the transactions table
			accumulateByTags(&is.Taxes, taxesTags, category+"/"+subCategory, -amount)
			accumulateByTags(&is.Expenses, expensesTags, category+"/"+subCategory, -amount)
		}
	}
	is.RevenueNetTaxes = is.Revenue.Total - is.Taxes.Total
	is.OperatingIncome = is.RevenueNetTaxes - is.Expenses.Total
	is.TotalEarnings = is.OperatingIncome + is.Investments.Total
	return
}

func accumulateByTags(
	rg *ReportGroup, matchers []string, tag string, amount int64,
) {
	if stringMatchList(tag, matchers) {
		rg.Total += amount
		for _, m := range matchers {
			if stringMatch(tag, m) {
				old := rg.Groups[m]
				rg.Groups[m] = old + amount
			}
		}
	}
}
