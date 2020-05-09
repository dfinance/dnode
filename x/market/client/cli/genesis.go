package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/dfinance/dnode/x/market/internal/types"
)

// AddMarketNomineesCmd return genesis tx command which adds nominees to genesis state.
func AddMarketNomineesCmd(ctx *server.Context, cdc *codec.Codec, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-market-nominees-gen [address1,address2...]",
		Short: "Add market nominees to genesis.json",
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

			// add nominee account to the app state
			var genesisMarket types.GenesisState

			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisMarket)
			var nominees []string
			for _, n := range addresses {
				found := false
				for _, gn := range genesisMarket.Params.Nominees {
					if n == gn {
						nominees = append(nominees, gn)
						found = true
						break
					}
				}
				if !found {
					nominees = append(nominees, n)
				}
			}

			genesisMarket.Params.Nominees = nominees

			// update the app state
			genesisStateBz := cdc.MustMarshalJSON(genesisMarket)
			appState[types.ModuleName] = genesisStateBz

			// export app state
			appStateJSON, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}
			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	return cmd
}
