package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// GetCalls returns query command that return calls.
func GetCalls(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "calls",
		Short: "Get active calls to confirm",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCalls), nil)
			if err != nil {
				return err
			}

			var out types.CallsResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCall returns query command that return call by callID.
func GetCall(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "call [id]",
		Short:   "Get call by ID",
		Example: "call 100",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			callID, err := helpers.ParseDnIDParam("id", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := types.CallReq{CallID: callID}
			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCall), bz)
			if err != nil {
				return err
			}

			var out types.CallResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"callID [uint]",
	})

	return cmd
}

// GetCallByUniqueID returns query command that return call by call uniqueID.
func GetCallByUniqueID(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unique [uniqueID]",
		Short:   "get call by unique id",
		Example: "unique issue1",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// prepare request
			req := types.CallByUniqueIdReq{UniqueID: args[0]}
			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCallByUnique), bz)
			if err != nil {
				return err
			}

			var out types.CallResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"call uniqueID [string]",
	})

	return cmd
}

// GetLastId returns query command that return last call ID.
func GetLastId(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "lastId",
		Short: "Get last call ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryLastId), nil)
			if err != nil {
				return err
			}

			var out types.LastCallIdResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
