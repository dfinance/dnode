package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/currencies_register/internal/keeper"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// GetCmdInfo returns query command that returns currencyInfo by denom.
func GetCmdInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "info [denom]",
		Example: "dncli currencies_register info dfi",
		Short:   "Get currency info by denom",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)

			// prepare request
			req := types.CurrencyInfoReq{Denom: args[0]}

			bz, err := ctx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryInfo), bz)
			if err != nil {
				return err
			}

			var out types.CurrencyInfo
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
}
