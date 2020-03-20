/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package commands

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/oracle-app/internal/app"
	"github.com/dfinance/dnode/oracle-app/internal/exchange"
)

// init configuration file
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration file initialized!")
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts an oracle application",
	// 	Long: `A longer description that spans multiple lines and likely contains examples
	// and usage of using your command. For example:
	//
	// Cobra is a CLI library for Go that empowers applications.
	// This application is a tool to generate the needed files
	// to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		level, err := logrus.ParseLevel(flagLogLevel)
		if err != nil {
			logger.Fatal(err)
		}
		logger.SetOutput(os.Stdout)
		logger.SetLevel(level)

		exchange.SetLogger(logger)

		assets := make(map[string][]exchange.Asset)
		err = viper.UnmarshalKey("exchanges", &assets)
		if err != nil {
			logger.Fatal(err)
		}
		app, err := app.NewOracleApp(&app.Config{
			ChainID:     flagChainID,
			Mnemonic:    viper.GetString("MNEMONIC"),
			Account:     viper.GetUint32("ACCOUNT"),
			Index:       viper.GetUint32("INDEX"),
			Passphrase:  viper.GetString("PASSPHRASE"),
			AccountName: viper.GetString("ACCNAME"),
			APIURL:      flagAPIURL,
			Gas:         flagGas,
			Fees:        flagFees,
			Logger:      logger,
			Assets:      assets,
		})
		if err != nil {
			logrus.Fatal(err)
		}
		if err := app.Start(); err != nil {
			logrus.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
