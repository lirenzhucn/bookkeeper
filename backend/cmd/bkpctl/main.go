package main

import (
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bkpctl/cmd"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

func main() {
	bookkeeper.SetupZapGlobals()
	cmd.Init()
	cmd.Execute()
}
