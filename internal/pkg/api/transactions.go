package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

var Transactions []bookkeeper.Transaction

func returnAllTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := bookkeeper.GetAllTransactions(dbpool, MAX_NUM_RECORDS, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
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
