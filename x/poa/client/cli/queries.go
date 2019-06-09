package cli

import (
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	"wings-blockchain/x/poa/queries"
	"github.com/cosmos/cosmos-sdk/codec"
	"fmt"
)

func GetValidators(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: 	"validators",
		Short:  "get validators list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/validators", queryRoute),
				nil)

			if err != nil {
				return err
			}

			var out queries.QueryValidatorsRes
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

func GetMinMax(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: 	"minmax",
		Short:  "get min/max values for validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/minmax", queryRoute),
				nil)

			if err != nil {
				return err
			}

			var out queries.QueryMinMaxRes
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}


func GetValidator(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: 	"validator [address]",
		Short:  "get validator by address",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/validators/%s", queryRoute, args[0]),
				nil)

			if err != nil {
				return err
			}

			var out queries.QueryGetValidatorRes
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
