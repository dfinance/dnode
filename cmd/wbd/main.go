package main

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccscli "github.com/cosmos/cosmos-sdk/x/genaccounts/client/cli"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/app"
	wbConfig "github.com/dfinance/dnode/cmd/config"
	oraclecli "github.com/dfinance/dnode/x/oracle/client/cli"
	poaCli "github.com/dfinance/dnode/x/poa/client/cli"
	vmCli "github.com/dfinance/dnode/x/vm/client/cli"
)

// WBD (Daemon) entry function.
func main() {
	config := sdk.GetConfig()
	wbConfig.InitBechPrefixes(config)
	config.Seal()

	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "wbd",
		Short:             "Wings blockchain app daemon (server).",
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
		oraclecli.AddOracleNomeneesCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		testnetCmd(ctx, cdc, app.ModuleBasics, genaccounts.AppModuleBasic{}),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "WB", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

// Creating new WB app.
func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	// read VM config
	config, err := wbConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		panic(err)
	}

	return app.NewWbServiceApp(logger, db, config)
}

// Exports genesis data and validators.
func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	config, err := wbConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	if err != nil {
		panic(err)
	}

	if height != -1 {
		wbApp := app.NewWbServiceApp(logger, db, config)
		err := wbApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return wbApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	wbApp := app.NewWbServiceApp(logger, db, config)
	return wbApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// Init cmd together with VM configruation.
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager,
	defaultNodeHome string) *cobra.Command { // nolint: golint
	cmd := genutilcli.InitCmd(ctx, cdc, mbm, defaultNodeHome)

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		wbConfig.ReadVMConfig(viper.GetString(cli.HomeFlag))
	}

	return cmd
}
