package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"

	"github.com/gorilla/mux"
)

var Accounts []bookkeeper.Account

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/accounts", returnAllAccounts)
	myRouter.HandleFunc("/accounts/{id}", returnSingleAccount)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func populateAccounts() {
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

func main() {
	populateAccounts()
	handleRequests()
}
