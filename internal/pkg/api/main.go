package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func HandleRequests(port string) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/accounts", returnAllAccounts)
	myRouter.HandleFunc("/accounts/{id}", returnSingleAccount)
	myRouter.HandleFunc("/transactions", returnAllTransactions)
	myRouter.HandleFunc("/transactions/{id}", returnSingleTransaction)
	err := http.ListenAndServe(":"+port, myRouter)
	if err != nil {
		sugar.Errorw("Web server failed", "error", err)
		os.Exit(1)
	}
}
