package cmd

import (
	"fmt"
	"log"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/api"
	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bkpsrv",
	Short: "The bookkeeper server",
	Long: `The bookkeeper server (bkpsrv) provides a RESTful API endpoint for
keeping financial records and getting various reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("running server on port %s", cmd.Flags().Lookup("port").Value.String())
		port, err := cmd.Flags().GetInt("port")
		cobra.CheckErr(err)
		api.HandleRequests(fmt.Sprintf("%d", port))
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func Init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.bkpsrv.yaml)")
	rootCmd.Flags().IntP("port", "p", 10000, "the port of the server")
}

func initConfig() {
	bookkeeper.InitConfigYaml(rootCmd, ".bkpsrv", cfgFile)
}
