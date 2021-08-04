package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

func getAccountBalanceOn(w http.ResponseWriter, r *http.Request) {
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
	account, balance, err := bookkeeper.ComputeAccountBalance(dbpool, accountName, date)
	if !checkErr(err, w, 500, "Failed to query account balance",
		"accountName", accountName) {
		return
	}
	account_ := bookkeeper.AccountWithBalance{Account: account, Balance: balance}
	json.NewEncoder(w).Encode(account_)
}
