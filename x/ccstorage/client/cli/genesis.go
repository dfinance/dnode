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
	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

// AddGenesisCurrencyInfo return genesis cmd which adds currency into node genesis state.
func AddGenesisCurrencyInfo(ctx *server.Context, cdc *codec.Codec, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-currency [denom] [decimals] [vmBalancePath] [vmInfoPath]",
		Short: "Set currency to genesis state (non-token)",
		Args:  cobra.ExactArgs(4),
		RunE: func(_ *cobra.Command, args []string) error {
			// setup viper config
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			denom := args[0]
			if err := helpers.ValidateDenomParam("denom", denom, helpers.ParamTypeCliArg); err != nil {
				return err
			}

			decimals, err := helpers.ParseUint8Param("decimals", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			balancePath, _, err := helpers.ParseHexStringParam("vmBalancePath", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			infoPath, _, err := helpers.ParseHexStringParam("vmInfoPath", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			genesisState := types.GenesisState{}
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisState)

			// update the state
			params := types.CurrencyParams{
				Denom:          denom,
				Decimals:       decimals,
				BalancePathHex: balancePath,
				InfoPathHex:    infoPath,
			}
			if err := params.Validate(); err != nil {
				return fmt.Errorf("invalid params: %w", err)
			}

			foundIdx := -1
			for idx, params := range genesisState.CurrenciesParams {
				if params.Denom == denom {
					foundIdx = idx
					break
				}
			}
			if foundIdx == -1 {
				genesisState.CurrenciesParams = append(genesisState.CurrenciesParams, params)
			} else {
				genesisState.CurrenciesParams[foundIdx] = params
			}

			// update and export app state
			genesisStateBz := cdc.MustMarshalJSON(genesisState)
			appState[types.ModuleName] = genesisStateBz

			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	helpers.BuildCmdHelp(cmd, []string{
		"currency denomination symbol",
		"currency decimals count",
		"DVM account balance path",
		"DVM CurrencyInfo path",
	})

	return cmd
}
