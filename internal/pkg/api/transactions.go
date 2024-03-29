package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/api/_peg"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"go.uber.org/zap"
)

func prepQueryData(queryData *_peg.QueryData) {
	for i := range queryData.Values {
		queryData.Clause = strings.Replace(queryData.Clause, "$$",
			fmt.Sprintf("$%d", i+1), 1)
	}
}

func queryTransactions(w http.ResponseWriter, r *http.Request) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	queryString := strings.Trim(r.FormValue("queryString"), "'")
	sugar.Infow("got a query string", "queryString", queryString)
	queryData, err := _peg.ParseString(queryString)
	if !checkErr(err, w, 400, "Invalid query string", "error", err) {
		return
	}
	prepQueryData(&queryData)
	transactions, err := bookkeeper.GetTransactionsWithFilters(
		dbpool, queryData.Clause, queryData.Values, MAX_NUM_RECORDS)
	if err != nil {
		sugar.Errorw(
			"failed to query transactions with filters",
			"error", err,
			"queryData.Clause", queryData.Clause,
			"queryData.Values", queryData.Values,
		)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	// write the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

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
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if !checkErr(err, w, 400, "Invalid transaction id provided") {
		return
	}
	sugar.Infow("received valid PATCH request", "id", id)
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
	if !trans.Validate() {
		checkErr(
			fmt.Errorf("validation of transaction failed"), w, 400,
			"Invalid account payload", "transaction", trans,
		)
		return
	}

	if transId < 0 {
		err = bookkeeper.InsertTransaction(dbpool, &trans)
	} else {
		err = bookkeeper.UpdateTransaction(dbpool, &trans)
	}
	if !checkErr(err, w, 500, "Failed to insert or update transaction") {
		return
	}
	json.NewEncoder(w).Encode(trans)
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
