package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	"fmt"
	"wings-blockchain/x/multisig/queries"
)

func GetLastId(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "lastId",
		Short: "get last id to confirm",
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/lastId", queryRoute), nil)
			if err != nil {
				return err
			}

			var out queries.QueryLastIdRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
