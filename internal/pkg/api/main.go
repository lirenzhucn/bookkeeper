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
	// home page
	myRouter.Path("/").
		Methods("GET").
		HandlerFunc(homePage)
	// accounts
	myRouter.Path("/accounts").
		Methods("GET").
		HandlerFunc(returnAllAccounts)
	myRouter.Path("/accounts/{id}").
		Methods("GET").
		HandlerFunc(returnSingleAccount)
	myRouter.Path("/accounts").
		Methods("POST").
		HandlerFunc(postAccount)
	myRouter.Path("/accounts/{id}").
		Methods("PATCH").
		HandlerFunc(patchAccount)
	myRouter.Path("/accounts/{id}").
		Methods("DELETE").
		HandlerFunc(deleteAccount)
	// transactions
	myRouter.Path("/transactions").
		Methods("GET").
		Queries("startDate", "{startDate}", "endDate", "{endDate}").
		HandlerFunc(returnTransactionsBetweenDates)
	myRouter.Path("/transactions").
		Methods("GET").
		HandlerFunc(returnAllTransactions)
	myRouter.Path("/transactions/{id}").
		Methods("GET").
		HandlerFunc(returnSingleTransaction)
	myRouter.Path("/transactions").
		Methods("POST").
		HandlerFunc(postTransaction)
	myRouter.Path("/transactions/{id}").
		Methods("PATCH").
		HandlerFunc(patchTransaction)
	myRouter.Path("/transactions/{id}").
		Methods("DELETE").
		HandlerFunc(deleteTransaction)
	// reporting
	myRouter.Path("/reporting/account_balance").
		Methods("GET").
		Queries("accountName", "{accountName}", "date", "{date}").
		HandlerFunc(getAccountBalanceOnDateByName)
	myRouter.Path("/reporting/account_balance").
		Methods("GET").
		Queries("date", "{date}").
		HandlerFunc(getAllAccountsBalanceOnDate)
	myRouter.Path("/reporting/balance_sheet").
		Methods("GET").
		Queries(
			"date", "{date}", "assetTags", "{assetTags}",
			"liabilityTags", "{liabilityTags}",
		).
		HandlerFunc(getBalanceSheet)
	myRouter.Path("/reporting/income_statement").
		Methods("GET").
		Queries(
			"dateRange", "{dateRange}", "revenueTags", "{revenueTags}",
			"taxesTags", "{taxesTags}", "expensesTags", "{expensesTags}",
			"investmentsTags", "{investmentsTags}",
		).
		HandlerFunc(getIncomeStatement)
	err := http.ListenAndServe(":"+port, myRouter)
	if err != nil {
		sugar.Errorw("Web server failed", "error", err)
		os.Exit(1)
	}
}
