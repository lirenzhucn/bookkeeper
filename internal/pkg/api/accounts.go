package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

var Accounts []bookkeeper.Account

func returnAllAccounts(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Accounts)
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
