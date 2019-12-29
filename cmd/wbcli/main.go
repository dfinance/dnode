package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"os"
	"path"
	wbConfig "wings-blockchain/cmd/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	app "wings-blockchain"
	ccClient "wings-blockchain/x/currencies/client"
	ccRoutes "wings-blockchain/x/currencies/client/rest"
	msClient "wings-blockchain/x/multisig/client"
	msRoutes "wings-blockchain/x/multisig/client/rest"
	poaClient "wings-blockchain/x/poa/client"
	poaRoutes "wings-blockchain/x/poa/client/rest"
)

const (
	storeAcc = "acc"
	storeCC  = "currencies"
	storePoa = "poa"
	storeMC  = "multisig"
)

var defaultCLIHome = os.ExpandEnv("$HOME/.wbcli")

type ModuleClient interface {
	GetQueryCmd() *cobra.Command
	GetTxCmd() *cobra.Command
}

func main() {
	config := sdk.GetConfig()
	wbConfig.InitBechPrefixes(config)
	config.Seal()

	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	mc := []ModuleClient{
		ccClient.NewModuleClient(storeCC, cdc),
		poaClient.NewModuleClient(storePoa, cdc),
		msClient.NewModuleClient(storeMC, cdc),
	}

	rootCmd := &cobra.Command{
		Use:   "wbcli",
		Short: "wings blockchain client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(defaultCLIHome),
		queryCmd(cdc, mc),
		txCmd(cdc, mc),
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		client.LineBreak,
		keys.Commands(),
		client.LineBreak,
	)

	executor := cli.PrepareMainCmd(rootCmd, "WB", defaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

// 0.36.0 Breaking Changes
// [\#4588](https://github.com/cosmos/cosmos-sdk/issues/4588) Context does not depend on x/auth anymore.
// client/context is stripped out of the following features:
//- GetAccountDecoder()
//- CLIContext.WithAccountDecoder()
//- CLIContext.WithAccountStore()
//x/auth.AccountDecoder is unnecessary and consequently removed.
func registerRoutes(rs *lcd.RestServer) {
	rpc.RegisterRPCRoutes(rs.CliCtx, rs.Mux)
	auth.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, storeAcc)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux)
	ccRoutes.RegisterRoutes(rs.CliCtx, rs.Mux, rs.CliCtx.Codec)
	msRoutes.RegisterRoutes(rs.CliCtx, rs.Mux, rs.CliCtx.Codec)
	poaRoutes.RegisterRoutes(rs.CliCtx, rs.Mux, rs.CliCtx.Codec)
}

func queryCmd(cdc *amino.Codec, mc []ModuleClient) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(cdc),
	)

	for _, m := range mc {
		queryCmd.AddCommand(m.GetQueryCmd())
	}

	return queryCmd
}

func txCmd(cdc *amino.Codec, mc []ModuleClient) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
