package main

import "github.com/lirenzhucn/bookkeeper/internal/pkg/api"

func main() {
	api.PopulateAccounts()
	api.HandleRequests("10000")
}
