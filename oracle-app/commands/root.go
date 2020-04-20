package commands

import (
	"fmt"
	"log"
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
	rootCmd.PersistentFlags().String("chain-id", "dn-testnet", "sets the chain ID")
	rootCmd.PersistentFlags().String("api-url", "http://127.0.0.1:1317", "sets an URL for API requests")
	rootCmd.PersistentFlags().String("fees", "1dfi", "sets the transaction fees")
	rootCmd.PersistentFlags().Uint64("gas", 200000, "sets the gas fees")

	if err := viper.BindPFlag("chain-id", rootCmd.PersistentFlags().Lookup("chain-id")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	if err := viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	if err := viper.BindPFlag("fees", rootCmd.PersistentFlags().Lookup("fees")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	if err := viper.BindPFlag("gas", rootCmd.PersistentFlags().Lookup("gas")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
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

		// Search config in home directory with name ".pricefeed-app" (without extension).
		viper.SetConfigType("yaml")
		viper.SetConfigFile(path.Join(home, ".pricefeed-app", "config.yaml"))
	}

	viper.SetEnvPrefix("DN_PRICEFEEDAPP")
	viper.AutomaticEnv() // read in environment variables that match

	viper.SafeWriteConfig()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Config file:", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}

	flagChainID = viper.GetString("chain-id")
	flagFees = viper.GetString("fees")
	flagGas = viper.GetUint64("gas")
	flagAPIURL = viper.GetString("api-url")
}
