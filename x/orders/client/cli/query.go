package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

const (
	flagOrderOwner     = "owner"
	flagOrderDirection = "direction"
	flagOrderMarketID  = "market-id"
)

// GetCmdListOrders returns query command that lists all order objects with filters and pagination.
func GetCmdListOrders(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Args:    cobra.ExactArgs(0),
		Example: "list",
		Short:   "Lists all orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			ownerFilterStr := viper.GetString(flagOrderOwner)
			directionFilterStr := viper.GetString(flagOrderDirection)
			marketIDFilter := viper.GetString(flagOrderMarketID)
			pageStr, limitStr := viper.GetString(flags.FlagPage), viper.GetString(flags.FlagLimit)
			page, limit, err := helpers.ParsePaginationParams(pageStr, limitStr, helpers.ParamTypeCliFlag)
			if err != nil {
				return err
			}

			ownerFilter := sdk.AccAddress{}
			if ownerFilterStr != "" {
				var err error
				ownerFilter, err = sdk.AccAddressFromBech32(ownerFilterStr)
				if err != nil {
					return fmt.Errorf("%s argument %q parse error: %w", flagOrderOwner, ownerFilterStr, err)
				}
			}

			// prepare request
			req := types.OrdersReq{
				Page:      page,
				Limit:     limit,
				Owner:     ownerFilter,
				Direction: types.NewDirectionRaw(directionFilterStr),
				MarketID:  marketIDFilter,
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

			var out types.Orders
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
	helpers.AddPaginationCmdFlags(cmd)
	cmd.Flags().String(flagOrderOwner, "", "(optional) filter by owner address")
	cmd.Flags().String(flagOrderDirection, "", "(optional) filter by direction (bid/ask)")
	cmd.Flags().String(flagOrderMarketID, "", "(optional) filter by marketID")

	return cmd
}

// GetCmdOrder returns query command that returns order by id.
func GetCmdOrder(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "order [id]",
		Example: "order 1",
		Short:   "Get order by id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			orderID, err := helpers.ParseDnIDParam("id", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := types.OrderReq{
				ID: orderID,
			}

			bz, err := ctx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryOrder), bz)
			if err != nil {
				return err
			}

			var out types.Order
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"order ID [uint]",
	})

	return cmd
}
