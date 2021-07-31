package bookkeeper

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

func InitConfigYaml(cmd *cobra.Command, configName string) {
	InitConfig(cmd, configName, "yaml")
}

func InitConfig(cmd *cobra.Command, configName string, configType string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find the home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		// Search config in home directory with name <configName>
		viper.AddConfigPath(home)
		viper.SetConfigType(configType)
		viper.SetConfigName(configName)
	}
	// Set environment variables first
	viper.AutomaticEnv()
	// Then use values in the config file
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
	// Manually bind flags
	bindFlags(cmd)
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
