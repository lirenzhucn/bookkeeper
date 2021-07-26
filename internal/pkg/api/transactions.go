package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

var Transactions []bookkeeper.Transaction

func returnAllTransactions(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Transactions)
}

func returnSingleTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	for _, transaction := range Transactions {
		if strconv.Itoa(transaction.Id) == key {
			json.NewEncoder(w).Encode(transaction)
			break
		}
	}
}

func PopulateTransactions() {
	Transactions = []bookkeeper.Transaction{
		{Id: 1, Date: time.Now(), Desc: "Transaction 1", Amount: 1.0, Type: "credit", Category: "Groceries", AccountId: 1},
		{Id: 2, Date: time.Now(), Desc: "Transaction 2", Amount: 100.0, Type: "debit", Category: "Groceries", AccountId: 1},
	}
}
