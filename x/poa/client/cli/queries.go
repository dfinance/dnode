package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"wings-blockchain/x/poa/types"
)

// Get validators list from CLI.
func GetValidators(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "get validators list, their amount and required confirmations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/validators", queryRoute),
				nil)

			if err != nil {
				return err
			}

			var out types.ValidatorsConfirmations
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// Get min/max values for validators amount.
func GetMinMax(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "minmax",
		Short: "get min/max values for validators amount",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/minmax", queryRoute),
				nil)

			if err != nil {
				return err
			}

			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// Get a validator by address.
func GetValidator(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [address]",
		Short: "get validator by address",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			address, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			query := types.QueryValidator{
				Address: address,
			}

			bz, err := cliCtx.Codec.MarshalJSON(query)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/validators", queryRoute),
				bz)
			if err != nil {
				return err
			}

			var out types.Validator
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
