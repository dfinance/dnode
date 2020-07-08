package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/poa/internal/types"
)

// GetValidators returns query command that returns validators list with some meta.
func GetValidators(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "validators",
		Short:   "Get PoA validators list and required confirmations count",
		Example: "validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidators), nil)
			if err != nil {
				return err
			}

			var out types.ValidatorsConfirmationsResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetValidator returns query command that returns validator object.
func GetValidator(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validator [address]",
		Short:   "get validator by address",
		Example: "validator {account}",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			sdkAddr, err := helpers.ParseSdkAddressParam("address", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := types.ValidatorReq{
				Address: sdkAddr,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidator), bz)
			if err != nil {
				return err
			}

			var out types.Validator
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"validator SDK address",
	})

	return cmd
}

// GetMinMax returns query command that returns module params.
func GetMinMax(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "minmax",
		Short: "Get min/max values for validators amount",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryMinMax), nil)
			if err != nil {
				return err
			}

			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
