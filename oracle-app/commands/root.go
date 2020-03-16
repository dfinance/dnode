package commands

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagCfgFile  string
	flagChainID  string
	flagFees     string
	flagGas      uint64
	flagLogLevel string
	flagAPIURL   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oracle-app",
	Short: "An application that receives prices from the Binance exchange and writes them to the network",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&flagCfgFile, "config", "", "config file (default is $HOME/.oracle-app.yaml)")
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "warn", "sets an application log level (trace, debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().StringVar(&flagChainID, "chain-id", "dn-testnet", "sets the chain ID")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "http://127.0.0.1:1317", "sets an URL for API requests")
	rootCmd.PersistentFlags().StringVar(&flagFees, "fees", "1dfi", "sets the transaction fees")
	rootCmd.PersistentFlags().Uint64Var(&flagGas, "gas", 50000, "sets the gas fees")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if flagCfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(flagCfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".oracle-app" (without extension).
		viper.AddConfigPath(path.Join(home, ".oracle-app"))

		viper.SetConfigName("config.yml")
	}
	viper.SetEnvPrefix("DN_ORACLEAPP")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
