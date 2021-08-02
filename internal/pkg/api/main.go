package api

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"go.uber.org/zap"
)

var dbpool *pgxpool.Pool

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

func createDbPool(db_url string) error {
	sugar := zap.L().Sugar()
	defer sugar.Sync()
	var err error
	sugar.Infow("connecting to db", "db_url", bookkeeper.MaskDbPassword(db_url))
	dbpool, err = pgxpool.Connect(context.Background(), db_url)
	if err != nil {
		sugar.Errorw("failed to obtain DB conn pool", "db_url", db_url)
		return err
	}
	return nil
}

func HandleRequests(port string, db_url string) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	if err := createDbPool(db_url); err != nil {
		sugar.Errorw("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// set up routes
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
