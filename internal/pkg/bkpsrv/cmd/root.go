package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bkpsrv",
	Short: "The bookkeeper server",
	Long: `The bookkeeper server (bkpsrv) provides a RESTful API endpoint for
keeping financial records and getting various reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("port=%s", cmd.Flags().Lookup("port").Value.String())
		if port, err := cmd.Flags().GetInt("port"); err == nil {
			api.HandleRequests(fmt.Sprintf("%d", port))
		} else {
			log.Fatal(
				fmt.Sprintf("PANIC: port is not specified correctly: %s", err),
			)
		}
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
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find the home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		// Search config in home directory with name ".bkpsrv"
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bkpsrv")
	}
	// Set environment variables first
	viper.AutomaticEnv()
	// Then use values in the config file
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
	// Manually bind flags
	bindFlags(rootCmd)
}

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Deal with environment variable names
		if strings.Contains(f.Name, "-") {
			envVar := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			viper.BindEnv(f.Name, envVar)
		}
		// Apply the viper config values to the flag when the flag is not set
		// and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
