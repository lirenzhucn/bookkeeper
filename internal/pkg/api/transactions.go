package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"go.uber.org/zap"
)

var Transactions []bookkeeper.Transaction

func returnTransactionsBetweenDates(w http.ResponseWriter, r *http.Request) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	startDateStr := r.FormValue("startDate")
	endDateStr := r.FormValue("endDate")
	// parse date string
	start, err := time.Parse("2006/01/02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid query term startDate", 400)
		return
	}
	end, err := time.Parse("2006/01/02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid query term endDate", 400)
		return
	}
	// move end time from the beginning of the day to the end of the day
	offset, _ := time.ParseDuration("23h59m59s")
	end = end.Add(offset)
	// query the database
	transactions, err := bookkeeper.GetTransactionsBetweenDates(
		dbpool, start, end, MAX_NUM_RECORDS)
	if err != nil {
		sugar.Errorw("failed to query transactions between two dates", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	// write the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

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
