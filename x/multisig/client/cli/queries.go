// Implements CLI queries multisig modules.
package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"strconv"
	"wings-blockchain/x/multisig/types"
)

// Get calls from CLI.
func GetCalls(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "calls",
		Short: "get active calls to confirm",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/calls", queryRoute), nil)

			if err != nil {
				return err
			}

			var out types.CallsResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get amount of calls from CLI.
func GetLastId(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "lastId",
		Short: "get last call id",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/lastId", queryRoute), nil)
			if err != nil {
				return err
			}

			var resp types.LastIdRes
			cdc.MustUnmarshalJSON(res, &resp)
			return cliCtx.PrintOutput(resp)
		},
	}
}

// Get call by id from CLI.
func GetCall(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "call [id]",
		Short: "get call by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			callId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			callReq := types.CallReq{CallId: callId}
			bz, err := cliCtx.Codec.MarshalJSON(callReq)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/call", queryRoute), bz)
			if err != nil {
				return err
			}

			var resp types.CallResp
			cdc.MustUnmarshalJSON(res, &resp)
			return cliCtx.PrintOutput(resp)
		},
	}
}

//Get call by unique id from CLI.
func GetCallByUniqueID(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unique [unique_id]",
		Short: "get call by unique id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			callReq := types.UniqueReq{UniqueId: args[0]}
			bz, err := cliCtx.Codec.MarshalJSON(callReq)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unique", queryRoute), bz)
			if err != nil {
				return err
			}

			var resp types.CallResp
			cdc.MustUnmarshalJSON(res, &resp)
			return cliCtx.PrintOutput(resp)
		},
	}
}
