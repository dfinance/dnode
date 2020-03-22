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
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccscli "github.com/cosmos/cosmos-sdk/x/genaccounts/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/app"
	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
	oracleCli "github.com/dfinance/dnode/x/oracle/client/cli"
	poaCli "github.com/dfinance/dnode/x/poa/client/cli"
	vmCli "github.com/dfinance/dnode/x/vm/client/cli"
)

// DNODE (Daemon) entry function.
func main() {
	config := sdk.GetConfig()
	dnConfig.InitBechPrefixes(config)
	config.Seal()

	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "dnode",
		Short:             "Dfinance blockchain app daemon (server).",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(
		InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{},
			genaccounts.AppModuleBasic{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccscli.AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		// Allows user to poa genesis validator
		poaCli.AddGenesisPoAValidatorCmd(ctx, cdc),
		vmCli.GenesisWSFromFile(ctx, cdc),
		oracleCli.AddOracleNomineesCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		oracleCli.AddAssetGenCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		testnetCmd(ctx, cdc, app.ModuleBasics, genaccounts.AppModuleBasic{}),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// configure Sentry integration
	if err := sentry.Init(helpers.GetSentryOptions(version.ServerName, version.Version, version.Commit)); err != nil {
		stdLog.Fatalf("sentry init: %v", err)
	}
	defer helpers.SentryDeferHandler()

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "DN", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		helpers.CrashWithError(err)
		helpers.CrashWithError(err)
	}
}

// Creating new DN app.
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	// read VM config
	config, err := dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		helpers.CrashWithError(err)
	}

	return app.NewDnServiceApp(logger, db, config)
}

// Exports genesis data and validators.
func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	config, err := dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		helpers.CrashWithError(err)
	}

	if height != -1 {
		dnApp := app.NewDnServiceApp(logger, db, config)
		err := dnApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return dnApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	dnApp := app.NewDnServiceApp(logger, db, config)
	return dnApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// Init cmd together with VM configruation.
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager,
	defaultNodeHome string) *cobra.Command { // nolint: golint
	cmd := genutilcli.InitCmd(ctx, cdc, mbm, defaultNodeHome)

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		dnConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	}

	return cmd
}
