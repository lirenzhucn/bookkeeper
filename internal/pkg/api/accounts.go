package api

import (
	"encoding/json"
	"fmt"
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

func PopulateAccounts() {
	Accounts = []bookkeeper.Account{
		{Id: 1, Name: "LZ Chase Checking", Type: "Debit", Category: "Asset", Balance: 0},
		{Id: 2, Name: "LZ Chase Ultimate Freedom", Type: "Credit", Category: "Liability", Balance: 0},
	}
	for _, account := range Accounts {
		if !account.Validate() {
			fmt.Printf("Invalid account %d\n", account.Id)
		}
	}
}
