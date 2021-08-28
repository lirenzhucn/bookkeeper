package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func returnAccountByName(w http.ResponseWriter, r *http.Request) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	accountName := r.FormValue("accountName")
	sugar.Infow("got a query on account", "accountName", accountName)
	account, err := bookkeeper.GetSingleAccountByName(dbpool, accountName)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, "Account not found", 404)
		} else {
			http.Error(w, "Internal server error", 500)
			sugar.Errorw("failed to get account", "accountName", accountName,
				"error", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}

func postAccount(w http.ResponseWriter, r *http.Request) {
	postOrPatchAccount(w, r, -1)
}

func patchAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if !checkErr(err, w, 400, "Invalid account id provided") {
		return
	}
	postOrPatchAccount(w, r, id)
}

func postOrPatchAccount(w http.ResponseWriter, r *http.Request, accountId int) {
	var account bookkeeper.Account

	body, err := ioutil.ReadAll(r.Body)
	if !checkErr(err, w, 400, "Failed to read the request body") {
		return
	}
	err = json.Unmarshal(body, &account)
	if !checkErr(err, w, 400, "Failed to parse the request body as a JSON string") {
		return
	}
	if !account.Validate() {
		checkErr(
			fmt.Errorf("validation of account failed"), w, 400,
			"Invalid account payload", "account", account,
		)
		return
	}

	if accountId < 0 {
		err := bookkeeper.InsertAccount(dbpool, &account)
		if !checkErr(err, w, 500, "Failed to insert account") {
			return
		}
		json.NewEncoder(w).Encode(account)
	} else {
		// overwrite the id in the payload
		account.Id = accountId
		err := bookkeeper.UpdateAccount(dbpool, &account)
		if err != nil && strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, "Cannot find account with the specified id", 404)
			return
		}
		if !checkErr(err, w, 500, "Failed to update account", "accout_id", accountId) {
			return
		}
		json.NewEncoder(w).Encode(account)
	}
}

func deleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if !checkErr(err, w, 400, "Invalid account id provided") {
		return
	}
	err = bookkeeper.DeleteAccount(dbpool, id)
	if err != nil && strings.Contains(err.Error(), "violates foreign key constraint") {
		http.Error(w, "Failed to update account. Account may be referenced by a transaction.", 500)
		return
	}
	if !checkErr(err, w, 500, "Failed to update account", "accout_id", id) {
		return
	}
}
