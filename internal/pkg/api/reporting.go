package api

import (
	"encoding/json"
	"net/http"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

func getAccountBalanceOnDateByName(w http.ResponseWriter, r *http.Request) {
	accountName := r.FormValue("accountName")
	date, ok := parseDateTimeInQueryAndFail(w, r, "date")
	if !ok {
		return
	}
	account, balance, err := bookkeeper.ComputeAccountBalanceByName(dbpool, accountName, date)
	if !checkErr(err, w, 500, "Failed to query account balance",
		"accountName", accountName) {
		return
	}
	account_ := bookkeeper.AccountWithBalance{Account: account, Balance: balance}
	json.NewEncoder(w).Encode(account_)
}

func getAllAccountsBalanceOnDate(w http.ResponseWriter, r *http.Request) {
	date, ok := parseDateTimeInQueryAndFail(w, r, "date")
	if !ok {
		return
	}
	accounts_, err := bookkeeper.GetAllAccountsBalanceOnDate(dbpool, date)
	if !checkErr(err, w, 500, "Failed to get the balance of all accounts",
		"error", err) {
		return
	}
	json.NewEncoder(w).Encode(accounts_)
}

func getBalanceSheet(w http.ResponseWriter, r *http.Request) {
	var balanceSheets []bookkeeper.BalanceSheet
	dates, ok := parseMultipleDateTimesInQueryAndFail(w, r, "date")
	if !ok {
		return
	}
	assetTags, ok := parseTagsInQueryAndFail(w, r, "assetTags")
	if !ok {
		return
	}
	liabilityTags, ok := parseTagsInQueryAndFail(w, r, "liabilityTags")
	if !ok {
		return
	}
	for _, date := range dates {
		accounts, err := bookkeeper.GetAllAccountsBalanceOnDate(dbpool, date)
		if !checkErr(err, w, 500, "Failed to get the balance of all accounts",
			"error", err) {
			return
		}
		balanceSheets = append(
			balanceSheets,
			bookkeeper.ComputeBalanceSheet(accounts, assetTags, liabilityTags),
		)
	}
	json.NewEncoder(w).Encode(balanceSheets)
}
