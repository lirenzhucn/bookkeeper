package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var dbConfigFile string
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Interact with database directly",
	Long: `db is for direct maintenance of the database without going through
the API server.`,
}
var dbTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to the database",
	Run: func(cmd *cobra.Command, args []string) {
		var msg string
		sugar := zap.L().Sugar()
		defer sugar.Sync()
		db_url, err := cmd.Flags().GetString("db-url")
		if err != nil {
			sugar.Errorw("get db_url failed", "error", err)
			os.Exit(1)
		}
		dbpool, err := pgxpool.Connect(context.Background(), db_url)
		if err != nil {
			sugar.Errorw("connection to database failed", "error", err)
		}
		err = dbpool.QueryRow(context.Background(), "select 'test'").Scan(&msg)
		if err != nil {
			sugar.Errorw("simply query to the database failed", "error", err)
		}
		fmt.Printf("Connection to database [%s] was successful\n",
			bookkeeper.MaskDbPassword(db_url))
	},
}
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database with the schema",
	Run:   dbInit,
}

func initDbCmd(rootCmd *cobra.Command) {
	cobra.OnInitialize(initDbConfig)
	dbCmd.PersistentFlags().StringVar(&dbConfigFile, "config", "",
		"config file (defualt is ./configs/.bkpctl_db.yaml)")
	dbCmd.PersistentFlags().StringP("db-url", "u", "", "Database URL")
	dbInitCmd.Flags().Bool("dry-run", false,
		"set this flag to print actions without taking them")
	dbInitCmd.Flags().StringP("data-file", "d", "",
		"path to initial data (default is empty)")
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbTestCmd)
	rootCmd.AddCommand(dbCmd)
}

func initDbConfig() {
	bookkeeper.InitConfigYaml(dbCmd, ".bkpctl_db", dbConfigFile)
}

func dbInit(cmd *cobra.Command, args []string) {
	db_url, err := cmd.Flags().GetString("db-url")
	cobra.CheckErr(err)
	fmt.Printf("Initializing the database at %s...\n",
		bookkeeper.MaskDbPassword(db_url))
	dryRun, err := cmd.Flags().GetBool("dry-run")
	cobra.CheckErr(err)
	dataFile, err := cmd.Flags().GetString("data-file")
	cobra.CheckErr(err)
	var dbpool *pgxpool.Pool = nil
	if !dryRun {
		dbpool, err = pgxpool.Connect(context.Background(), db_url)
		cobra.CheckErr(err)
	}
	commands, err := bookkeeper.InitDb(dbpool, dataFile, dryRun)
	cobra.CheckErr(err)
	if dryRun {
		fmt.Println(
			"Dry run is on. The following commands would have been executed:",
		)
		fmt.Println(">>>")
		for _, c := range commands {
			fmt.Println(c)
		}
		fmt.Println("<<<")
	}
}
