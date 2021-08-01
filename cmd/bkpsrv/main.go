package main

import (
	"log"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bkpsrv/cmd"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)
	cmd.Init()
	cmd.Execute()
}
