package cmd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var dbConfigFile string
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Interact with database directly",
	Long: `db is for direct maintenance of the database without going through
the API server.`,
}
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database with the schema",
	Run:   dbInit,
}

func initDbCmd(rootCmd *cobra.Command) {
	cobra.OnInitialize(initDbConfig)
	dbCmd.PersistentFlags().StringVar(&dbConfigFile, "config", "",
		"config file (defualt is $HOME/.bkpctl_db.yaml)")
	dbCmd.PersistentFlags().StringP("db-url", "u", "", "Database URL")
	dbInitCmd.Flags().Bool("dry-run", false,
		"set this flag to print actions without taking them")
	dbCmd.AddCommand(dbInitCmd)
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
	var dbpool *pgxpool.Pool = nil
	if !dryRun {
		dbpool, err = pgxpool.Connect(context.Background(), db_url)
		cobra.CheckErr(err)
	}
	commands, err := bookkeeper.InitDb(dbpool, dryRun)
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
