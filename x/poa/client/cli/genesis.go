package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/poa/internal/types"
)

// AddGenesisPoAValidatorCmd return genesis cmd which adds validator into node genesis state.
func AddGenesisPoAValidatorCmd(ctx *server.Context, cdc *codec.Codec, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-poa-validator [address] [ethAddress]",
		Short: "Adds PoA validator to genesis state",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// setup viper config
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			sdkAddr, err := helpers.ParseSdkAddressParam("address", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			ethAddr, err := helpers.ParseEthereumAddressParam("ethAddress", args[1], helpers.ParamTypeCliArg)
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

			// update and validate the state
			genesisState.Validators = append(genesisState.Validators, types.NewValidator(sdkAddr, ethAddr))
			if err := genesisState.Validate(true); err != nil {
				return err
			}

			// update and export app state
			genesisStateBz := cdc.MustMarshalJSON(genesisState)
			appState[types.ModuleName] = genesisStateBz

			appStateJson, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJson

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	helpers.BuildCmdHelp(cmd, []string{
		"validator SDK address",
		"validator Ethereum address",
	})

	return cmd
}
