package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

const (
	flagMarketBaseDenom  = "base-asset-denom"
	flagMarketQuoteDenom = "quote-asset-denom"
)

// GetCmdListMarkets returns query command that lists all market objects with filters and pagination.
func GetCmdListMarkets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.ExactArgs(0),
		Example: "dncli markets list --page=1 --limit=10 --base-asset-denom=btc",
		Short:   "Lists all markets by limit and page",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			baseDenomFilter := viper.GetString(flagMarketBaseDenom)
			quoteDenomFilter := viper.GetString(flagMarketQuoteDenom)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			// prepare request
			req := types.MarketsReq{
				Page:            page,
				Limit:           limit,
				BaseAssetDenom:  baseDenomFilter,
				QuoteAssetDenom: quoteDenomFilter,
			}

			bz, err := ctx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryList), bz)
			if err != nil {
				return err
			}

			var out types.Markets
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of markets to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of markets to query for")
	cmd.Flags().String(flagMarketBaseDenom, "", "(optional) filter by baseAsset denom")
	cmd.Flags().String(flagMarketQuoteDenom, "", "(optional) filter by quoteAsset denom")

	return cmd
}

// GetCmdMarket returns query command that returns market by id.
func GetCmdMarket(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "market [id]",
		Example: "dncli markets market 1",
		Short:   "Get market by id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			id, err := dnTypes.NewIDFromString(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q parse error: %w", "id", args[0], err)
			}

			// prepare request
			req := types.MarketReq{
				ID: id,
			}

			bz, err := ctx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryMarket), bz)
			if err != nil {
				return err
			}

			var out types.Market
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
}
