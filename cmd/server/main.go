package main

import "github.com/lirenzhucn/bookkeeper/internal/pkg/bkpsrv/cmd"

func main() {
	cmd.Init()
	cmd.Execute()
	// api.HandleRequests("10000")
}
