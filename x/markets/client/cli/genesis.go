package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	ccrTypes "github.com/dfinance/dnode/x/currencies_register"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// AddMarketGenCmd adds market to app genesis state.
func AddMarketGenCmd(ctx *server.Context, cdc *codec.Codec, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-market-gen [base_denom] [quote_denom]",
		Short: "Add market to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			baseDenom, quoteDenom := args[0], args[1]

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// retrive the currencies_register genesis
			var genesisCCRegister ccrTypes.GenesisState
			cdc.MustUnmarshalJSON(appState[ccrTypes.ModuleName], &genesisCCRegister)

			// retrieve the markets genesis
			var genesisMarket types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisMarket)

			// check if base/quote denom do exist in currencies_register genesis
			baseFound, quoteFound := false, false
			for _, ccInfo := range genesisCCRegister.Currencies {
				if ccInfo.Denom == baseDenom {
					baseFound = true
					continue
				}
				if ccInfo.Denom == quoteDenom {
					quoteFound = true
					continue
				}
			}
			if !baseFound {
				return fmt.Errorf("base asset denom currency not registered")
			}
			if !quoteFound {
				return fmt.Errorf("quote asset denom currency not registered")
			}

			// add market to the genesis
			marketID := len(genesisMarket.Params.Markets)

			genesisMarket.Params.Markets = append(genesisMarket.Params.Markets, types.Market{
				ID:              dnTypes.NewIDFromUint64(uint64(marketID)),
				BaseAssetDenom:  baseDenom,
				QuoteAssetDenom: quoteDenom,
			})

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
