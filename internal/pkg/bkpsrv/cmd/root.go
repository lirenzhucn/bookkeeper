package cmd

import (
	"fmt"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/api"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bkpsrv",
	Short: "The bookkeeper server",
	Long: `The bookkeeper server (bkpsrv) provides a RESTful API endpoint for
keeping financial records and getting various reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		sugar := zap.L().Sugar()
		defer sugar.Sync()
		sugar.Infow("running server",
			"port", cmd.Flags().Lookup("port").Value.String())
		port, err := cmd.Flags().GetInt("port")
		cobra.CheckErr(err)
		db_url, err := cmd.Flags().GetString("db-url")
		cobra.CheckErr(err)
		api.HandleRequests(fmt.Sprintf("%d", port), db_url)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func Init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&cfgFile, "config", "",
		"config file (default is ./configs/private/.bkpsrv.yaml)")
	rootCmd.Flags().IntP("port", "p", 10000, "the port of the server")
	rootCmd.Flags().StringP("db-url", "d", "", "URL to the database service")
}

func initConfig() {
	bookkeeper.InitConfigYaml(rootCmd, ".bkpsrv", cfgFile)
}
