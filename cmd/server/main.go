package main

import "github.com/lirenzhucn/bookkeeper/internal/pkg/api"

func main() {
	api.PopulateAccounts()
	api.PopulateTransactions()
	api.HandleRequests("10000")
}
