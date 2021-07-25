package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func HandleRequests(port string) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/accounts", returnAllAccounts)
	myRouter.HandleFunc("/accounts/{id}", returnSingleAccount)
	log.Fatal(http.ListenAndServe(":"+port, myRouter))
}
