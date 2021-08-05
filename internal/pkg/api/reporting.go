package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

func getAccountBalanceOnDateByName(w http.ResponseWriter, r *http.Request) {
	accountName := r.FormValue("accountName")
	dateStr := r.FormValue("date")
	if accountName == "" || dateStr == "" {
		http.Error(w, "Invalid query string", 400)
		return
	}
	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		http.Error(w, "Invalid query term date", 400)
		return
	}
	offset, _ := time.ParseDuration("23h59m59s")
	date = date.Add(offset)
	account, balance, err := bookkeeper.ComputeAccountBalanceByName(dbpool, accountName, date)
	if !checkErr(err, w, 500, "Failed to query account balance",
		"accountName", accountName) {
		return
	}
	account_ := bookkeeper.AccountWithBalance{Account: account, Balance: balance}
	json.NewEncoder(w).Encode(account_)
}

func getAllAccountsBalanceOnDate(w http.ResponseWriter, r *http.Request) {
	var accounts_ []bookkeeper.AccountWithBalance
	dateStr := r.FormValue("date")
	if dateStr == "" {
		http.Error(w, "Invalid query string", 400)
		return
	}
	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		http.Error(w, "Invalid query term date", 400)
		return
	}
	offset, _ := time.ParseDuration("23h59m59s")
	date = date.Add(offset)
	ids, err := bookkeeper.GetAllAccountIds(dbpool)
	if !checkErr(err, w, 500, "Failed to get all account ids") {
		return
	}
	for _, id := range ids {
		account, balance, err := bookkeeper.ComputeAccountBalanceById(dbpool, id, date)
		if !checkErr(err, w, 500,
			fmt.Sprintf("Failed to get balance for account id %d", id)) {
			return
		}
		accounts_ = append(
			accounts_,
			bookkeeper.AccountWithBalance{Account: account, Balance: balance},
		)
	}
	json.NewEncoder(w).Encode(accounts_)
}
