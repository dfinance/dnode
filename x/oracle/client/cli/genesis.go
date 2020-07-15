package cli

import (
	"fmt"
	"github.com/dfinance/dnode/helpers"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

const (
	flagClientHome = "home-client"
)

// AddOracleNomineesCmd returns add-oracle-nominees cobra Command for adding a nominee to genesis.
func AddOracleNomineesCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-nominees-gen [nomineeAddresses]",
		Short: "Add oracle nominees to genesis.json",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			addresses := strings.Split(args[0], ",")
			for i, a := range addresses {
				if _, err := sdk.AccAddressFromBech32(a); err != nil {
					return fmt.Errorf("%q address at index %d: %w", a, i, err)
				}
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// add genesis account to the app state
			var genesisOracle types.GenesisState

			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisOracle)
			var addedNomenees []string
			for _, n := range addresses {
				var found bool
				for _, nn := range genesisOracle.Params.Nominees {
					if n == nn {
						addedNomenees = append(addedNomenees, nn)
						found = true
					}
				}
				if !found {
					addedNomenees = append(addedNomenees, n)
				}
			}

			genesisOracle.Params.Nominees = addedNomenees

			genesisStateBz := cdc.MustMarshalJSON(genesisOracle)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"nomineeAddresses [string] comma separated list of nominee addresses",
	})

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	return cmd
}

// AddAssetGenCmd returns add-asset cobra Command for adding an asset to genesis.
func AddAssetGenCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-asset-gen [assetCode] [oracleAddresses]",
		Short: "Add oracle asset to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := types.ParseOracles(args[1])
			if err != nil {
				return fmt.Errorf("%s argument %q: %v", "oracles", args[1], err)
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracles", args[1])
			}

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// retrieve the module genesis state
			var genesisOracle types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisOracle)

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

			// update app state
			genesisStateBz := cdc.MustMarshalJSON(genesisOracle)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			// export app state
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"assetCode [string] asset code symbol",
		"oracleAddresses [string] comma separated list of oracle addresses",
	})

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}
