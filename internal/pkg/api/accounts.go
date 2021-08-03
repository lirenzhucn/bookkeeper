package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"go.uber.org/zap"
)

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
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	vars := mux.Vars(r)
	key := vars["id"]
	id, err := strconv.Atoi(key)
	if err != nil {
		http.Error(w, "Invalid id in query", 400)
		return
	}
	account, err := bookkeeper.GetSingleAccount(dbpool, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, "Account not found", 404)
		} else {
			http.Error(w, "Internal server error", 500)
			sugar.Errorw("failed to get account", "account_id", id, "error", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}
