package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"go.uber.org/zap"
)

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
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	vars := mux.Vars(r)
	key := vars["id"]
	id, err := strconv.Atoi(key)
	if err != nil {
		http.Error(w, "Invalid id in query", 400)
		return
	}
	transaction, err := bookkeeper.GetSingleTransaction(dbpool, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, "Transaction not found", 404)
		} else {
			http.Error(w, "Internal server error", 500)
			sugar.Errorw("failed to get transaction", "transaction_id", id, "error", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaction)
}

func postTransaction(w http.ResponseWriter, r *http.Request) {
	postOrPatchTransaction(w, r, -1)
}

func patchTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if !checkErr(err, w, 400, "Invalid transaction id provided") {
		return
	}
	postOrPatchTransaction(w, r, id)
}

func postOrPatchTransaction(w http.ResponseWriter, r *http.Request, transId int) {
	var trans bookkeeper.Transaction

	body, err := ioutil.ReadAll(r.Body)
	if !checkErr(err, w, 400, "Failed to read the request body") {
		return
	}
	err = json.Unmarshal(body, &trans)
	if !checkErr(err, w, 400, "Failed to parse the request body") {
		return
	}

	if transId < 0 {
		err := bookkeeper.InsertTransaction(dbpool, &trans)
		if !checkErr(err, w, 500, "Failed to insert account") {
			return
		}
		json.NewEncoder(w).Encode(trans)
	} else {
		// err := bookkeeper.UpdateTransaction(dbpool, &trans)
		http.Error(w, "PATCH not implemented", 404)
		return
	}
}

func deleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if !checkErr(err, w, 400, "Invalid account id provided") {
		return
	}
	err = bookkeeper.DeleteTransaction(dbpool, id)
	if !checkErr(err, w, 500, "Failed to update account", "accout_id", id) {
		return
	}
}
