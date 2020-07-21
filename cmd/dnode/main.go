package main

import (
	"encoding/json"
	"io"
	stdLog "log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	genutilCli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmTypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/app"
	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/logger"
	ccsCli "github.com/dfinance/dnode/x/ccstorage/client/cli"
	"github.com/dfinance/dnode/x/genaccounts"
	genaccsCli "github.com/dfinance/dnode/x/genaccounts/client/cli"
	marketsCli "github.com/dfinance/dnode/x/markets/client/cli"
	oracleCli "github.com/dfinance/dnode/x/oracle/client/cli"
	poaCli "github.com/dfinance/dnode/x/poa/client/cli"
	vmCli "github.com/dfinance/dnode/x/vm/client/cli"
)

// @title Dfinance dnode REST API
// @version 1.0

// @host localhost:1317
// @BasePath /
// @schemes http
// @query.collection.format multi

// DNODE (Daemon) entry function.
func main() {
	config := sdk.GetConfig()
	dnConfig.InitBechPrefixes(config)
	config.Seal()

	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	ctx.Logger = logger.NewDNLogger()

	rootCmd := &cobra.Command{
		Use:               "dnode",
		Short:             "Dfinance blockchain app daemon (server).",
		PersistentPreRunE: PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(
		InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		genutilCli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, app.DefaultNodeHome),
		genutilCli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
			genaccounts.AppModuleBasic{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
		genutilCli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccsCli.AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		// Allows user to poa genesis validator
		ccsCli.AddGenesisCurrencyInfo(ctx, cdc, app.DefaultNodeHome),
		poaCli.AddGenesisPoAValidatorCmd(ctx, cdc, app.DefaultNodeHome),
		vmCli.AddGenesisWSFromFileCmd(ctx, cdc),
		oracleCli.AddOracleNomineesCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		oracleCli.AddAssetGenCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		marketsCli.AddMarketGenCmd(ctx, cdc, app.DefaultNodeHome),
		testnetCmd(ctx, cdc, app.ModuleBasics, genaccounts.AppModuleBasic{}),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// configure crash logging
	if err := logger.SetupSentry(version.ServerName, version.Version, version.Commit); err != nil {
		stdLog.Fatal(err)
	}
	defer logger.CrashDeferHandler()

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "DN", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

// Creating new DN app.
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	// read VM config
	config, err := dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		panic(err)
	}

	return app.NewDnServiceApp(logger, db, config, dnConfig.DefInvCheckPeriod)
}

// Exports genesis data and validators.
func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmTypes.GenesisValidator, error) {
	config, err := dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		panic(err)
	}

	if height != -1 {
		dnApp := app.NewDnServiceApp(logger, db, config, dnConfig.DefInvCheckPeriod)
		err := dnApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return dnApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	dnApp := app.NewDnServiceApp(logger, db, config, dnConfig.DefInvCheckPeriod)
	return dnApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// Init cmd together with VM configruation.
// nolint
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager, defaultNodeHome string) *cobra.Command { // nolint: golint
	cmd := dnConfig.InitCmd(ctx, cdc, mbm, defaultNodeHome)

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	}

	return cmd
}
