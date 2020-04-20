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

	"github.com/dfinance/dnode/x/pricefeed/internal/types"
)

const (
	flagClientHome = "home-client"
)

// AddOracleNomineesCmd returns add-price-feed-nomenees cobra Command.
func AddPriceFeedNomineesCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-pricefeed-nominees-gen [address1,address2...]",
		Short: "Add price feed nominees to genesis.json",
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
			var genesisPriceFeed types.GenesisState

			cdc.MustUnmarshalJSON(appState[types.ModuleName], &genesisPriceFeed)
			var addedNomenees []string
			for _, n := range addresses {
				var found bool
				for _, nn := range genesisPriceFeed.Params.Nominees {
					if n == nn {
						addedNomenees = append(addedNomenees, nn)
						found = true
					}
				}
				if !found {
					addedNomenees = append(addedNomenees, n)
				}
			}

			genesisPriceFeed.Params.Nominees = addedNomenees

			genesisStateBz := cdc.MustMarshalJSON(genesisPriceFeed)
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

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")
	return cmd
}

// AddAssetGenCmd returns add-asset cobra Command.
func AddAssetGenCmd(ctx *server.Context, cdc *codec.Codec,
	defaultNodeHome, defaultClientHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-pricefeed-asset-gen [denom] [pricefeeds]",
		Short: "Add price feed for asset to genesis.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			denom := args[0]
			if len(denom) == 0 {
				return fmt.Errorf("%s argument %q: empty", "denom", args[0])
			}

			pricefeedArgs := strings.Split(args[1], ",")
			if len(pricefeedArgs) == 0 {
				return fmt.Errorf("%s argument: empty slice", "pricefeeds")
			}

			pricefeeds := make(types.PriceFeeds, 0, len(pricefeedArgs))
			for i, arg := range pricefeedArgs {
				addr, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return fmt.Errorf("%s argument: %q address at index %d: %w", "pricefeeds", arg, i, err)
				}
				pricefeeds = append(pricefeeds, types.PriceFeed{Address: addr})
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
				if asset.AssetCode == denom {
					foundIdx = i
					break
				}
			}

			if foundIdx == -1 {
				genesisOracle.Params.Assets = append(genesisOracle.Params.Assets, types.NewAsset(denom, pricefeeds, true))
			} else {
				genesisOracle.Params.Assets[foundIdx].PriceFeeds = pricefeeds
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

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, defaultClientHome, "client's home directory")

	return cmd
}
