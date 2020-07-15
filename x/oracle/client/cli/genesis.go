package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/helpers"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

const (
	flagClientHome = "home-client"
)

// AddOracleNomineesCmd returns add-oracle-nominees command for adding a nominee to genesis.
func AddOracleNomineesCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-nominees-gen [nomineeAddresses]",
		Short: "Add oracle nominees to genesis.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// setup viper config
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			addresses, err := helpers.ParseSdkAddressesParams("nomineeAddresses", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			var genesisOracle types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisOracle)

			// add genesis account to the app state
			var addedNominees []string
			for _, n := range addresses {
				var found bool
				for _, nn := range genesisOracle.Params.Nominees {
					if n.String() == nn {
						addedNominees = append(addedNominees, nn)
						found = true
					}
				}
				if !found {
					addedNominees = append(addedNominees, n.String())
				}
			}
			genesisOracle.Params.Nominees = addedNominees

			// update and export app state
			genesisStateBz := cdc.MustMarshalJSON(genesisOracle)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeAddresses comma separated list of nominee addresses",
	})
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}

// AddAssetGenCmd returns add-asset command for adding an asset to genesis.
func AddAssetGenCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-asset-gen [assetCode] [oracleAddresses]",
		Short: "Add oracle asset to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// setup viper config
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := parseOraclesArg("oracleAddresses", args[1])
			if err != nil {
				return err
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracleAddresses", args[1])
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			var genesisOracle types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisOracle)

			// add asset to the module state
			foundIdx := -1
			for i, asset := range genesisOracle.Params.Assets {
				if asset.AssetCode == assetCode {
					foundIdx = i
					break
				}
			}

			if foundIdx == -1 {
				genesisOracle.Params.Assets = append(genesisOracle.Params.Assets, types.NewAsset(assetCode, oracles, true))
			} else {
				genesisOracle.Params.Assets[foundIdx].Oracles = oracles
			}

			// update and export app state
			genesisStateBz := cdc.MustMarshalJSON(genesisOracle)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"comma separated list of oracle addresses",
	})
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}

// parseOraclesArg parses coma-separated notation oracle addresses and returns Oracles objects.
func parseOraclesArg(argName, argValue string) (retOracles types.Oracles, retErr error) {
	addresses, err := helpers.ParseSdkAddressesParams(argName, argValue, helpers.ParamTypeCliArg)
	if err != nil {
		retErr = err
		return
	}

	for _, address := range addresses {
		retOracles = append(retOracles, types.NewOracle(address))
	}

	return
}
