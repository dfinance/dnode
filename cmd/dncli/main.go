package main

import (
	"fmt"
	stdLog "log"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/app"
	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/logger"
	"github.com/dfinance/dnode/helpers/swagger"
	vmauthCli "github.com/dfinance/dnode/x/vmauth/client/cli"
)

const (
	// Default gas for CLI.
	DefaultGas = 500000
)

// Entry function for DN CLI.
func main() {
	config := sdk.GetConfig()
	dnConfig.InitBechPrefixes(config)
	config.Seal()

	// Set default gas.
	flags.GasFlagVar = flags.GasSetting{Gas: DefaultGas}

	cobra.EnableCommandSorting = false
	cdc := app.MakeCodec()

	rootCmd := &cobra.Command{
		Use:   "dncli",
		Short: "Dfinance blockchain client tool.",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		dnConfig.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		flags.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// configure crash logging
	if err := logger.SetupSentry(version.ClientName, version.Version, version.Commit); err != nil {
		stdLog.Fatal(err)
	}
	defer logger.CrashDeferHandler()

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "DN", app.DefaultCLIHome)
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}

// Registering routes for REST api.
func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
	swagger.RegisterRESTRoute(rs.Mux)
}

// Add query subcommands to CLI.
func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		vmauthCli.GetAccountCmd(cdc),
		flags.LineBreak,
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)
	DisableCommands(queryCmd, dnConfig.GetAppRestrictions().DisabledQueryCmd)

	return queryCmd
}

// Set default gas for tx commands.
func SetDefaultFeeForTxCmd(cmd *cobra.Command) {
	if feesFlag := cmd.Flag(flags.FlagFees); feesFlag != nil {
		feesFlag.DefValue = dnConfig.DefaultFee
		feesFlag.Usage = "Fees to pay along with transaction; eg: " + dnConfig.DefaultFee

		if feesFlag.Value.String() == "" {
			err := feesFlag.Value.Set(dnConfig.DefaultFee)
			if err != nil {
				panic(err)
			}
		}

		cmd.PreRunE = func(_ *cobra.Command, _ []string) error {
			coin, err := sdk.ParseCoin(feesFlag.Value.String())
			if err != nil {
				return fmt.Errorf("can't parse fees value to coins: %v", err)
			}

			if coin.Denom != dnConfig.MainDenom {
				return fmt.Errorf("fees must be paid only in %q, current fees in are %q", dnConfig.MainDenom, coin.Denom)
			}

			return nil
		}
	}
}

// DisableCommands disables cli commands.
func DisableCommands(cmd *cobra.Command, restrictedCmdList []string) {
	for _, child := range cmd.Commands() {
		for _, deniedCmd := range restrictedCmdList {
			if deniedCmd == child.Use {
				cmd.RemoveCommand(child)
				continue
			}
		}

		DisableCommands(child, restrictedCmdList)
	}
}

// Add transactions subcommands to CLI.
func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		flags.LineBreak,
		GetUpgradeTxCmd(cdc),
	)

	app.ModuleBasics.AddTxCommands(txCmd, cdc)
	SetDefaultFeeForTxCmd(txCmd)
	DisableCommands(txCmd, dnConfig.GetAppRestrictions().DisabledTxCmd)

	return txCmd
}

// Initialize CLI config.
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
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}

	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
