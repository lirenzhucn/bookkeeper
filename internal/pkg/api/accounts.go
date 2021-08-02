package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

var Accounts []bookkeeper.Account
var MAX_NUM_RECORDS int = 1000

func returnAllAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := bookkeeper.GetAllAccounts(dbpool, MAX_NUM_RECORDS, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

func returnSingleAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	for _, account := range Accounts {
		if strconv.Itoa(account.Id) == key {
			json.NewEncoder(w).Encode(account)
			break
		}
	}
}
