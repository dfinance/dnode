package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/market/internal/keeper"
	"github.com/dfinance/dnode/x/market/internal/types"
)

func GetCmdListMarkets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists all markets",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryList), nil)
			if err != nil {
				return err
			}

			var out types.Markets
			cdc.MustUnmarshalJSON(res, &out)

			return ctx.PrintOutput(out)
		},
	}
}
