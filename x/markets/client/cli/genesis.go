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
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// AddMarketGenCmd adds market to app genesis state.
func AddMarketGenCmd(ctx *server.Context, cdc *codec.Codec, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-market-gen [base_denom] [quote_denom]",
		Short:   "Add market to genesis.json",
		Example: "add-market-gen dfi eth",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			// parse inputs
			baseDenom, quoteDenom := args[0], args[1]
			if err := helpers.ValidateDenomParam("base_denom", baseDenom, helpers.ParamTypeCliArg); err != nil {
				return err
			}
			if err := helpers.ValidateDenomParam("quote_denom", quoteDenom, helpers.ParamTypeCliArg); err != nil {
				return err
			}

			// retrieve the app state
			genFile := config.GenesisFile()
			appState, genDoc, err := genutil.GenesisStateFromGenFile(cdc, genFile)
			if err != nil {
				return err
			}

			// retrieve the genesis
			var genesisCCS ccstorage.GenesisState
			cdc.MustUnmarshalJSON(appState[ccstorage.ModuleName], &genesisCCS)

			// retrieve the markets genesis
			var genesisMarket types.GenesisState
			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisMarket)

			// check if base/quote denom do exist in currencies genesis
			baseFound, quoteFound := false, false
			for _, params := range genesisCCS.CurrenciesParams {
				denom := params.Denom
				if denom == baseDenom {
					baseFound = true
					continue
				}
				if denom == quoteDenom {
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
			var id dnTypes.ID
			marketID := genesisMarket.LastMarketID
			if marketID == nil {
				id = dnTypes.NewZeroID()
			} else {
				id = marketID.Incr()
			}
			marketID = &id

			genesisMarket.LastMarketID = marketID
			genesisMarket.Markets = append(genesisMarket.Markets, types.Market{
				ID:              *marketID,
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
	helpers.BuildCmdHelp(cmd, []string{
		"base currency denomination symbol",
		"quote currency denomination symbol",
	})
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")

	return cmd
}
