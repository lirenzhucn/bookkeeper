package bookkeeper

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitConfigYaml(cmd *cobra.Command, configName string, cfgFile string) {
	InitConfig(cmd, configName, "yaml", cfgFile)
}

func InitConfig(cmd *cobra.Command, configName string, configType string, cfgFile string) {
	sugar := zap.L().Sugar()
	defer sugar.Sync()
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
		sugar.Infow("Use config file", "configFile", viper.ConfigFileUsed())
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
	// TODO: refactor this duplicated code!
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		// Deal with environment variable names
		if strings.Contains(f.Name, "-") {
			envVar := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			viper.BindEnv(f.Name, envVar)
		}
		// Apply the viper config values to the flag when the flag is not set
		// and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.PersistentFlags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
