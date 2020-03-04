// Querier commands currency module implementation for CLI.
package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/WingsDao/wings-blockchain/x/currencies/types"
)

// Get destroys by page & limit.
func GetDestroys(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "destroys [page] [limit]",
		Args:  cobra.ExactArgs(2),
		Short: "get destroys list by limit and page",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			page, isOk := sdk.NewIntFromString(args[0])
			if !isOk {
				return fmt.Errorf("%s argument %q is not a number, can't parse int", "page", args[0])
			}

			limit, isOk := sdk.NewIntFromString(args[1])
			if !isOk {
				return fmt.Errorf("%s argument %q is not a number, can't parse int", "limit", args[1])
			}

			req := types.DestroysReq{
				Page:  page,
				Limit: limit,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroys", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.Destroys
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get destroy by destroy id.
func GetDestroy(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy [destroyID]",
		Short: "get destroy by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			destroyId, isOk := sdk.NewIntFromString(args[0])
			if !isOk {
				return fmt.Errorf("%s argument %q is not a number, can't parse int", "destroyID", args[0])
			}

			req := types.DestroyReq{
				DestroyId: destroyId,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroy", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.Destroy
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get issue by issue id.
func GetIssue(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "issue [issueID]",
		Short: "get issue by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			req := types.IssueReq{IssueID: args[0]}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue", queryRoute), bz)
			if err != nil {
				return err
			}

			var out types.Issue
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get currency by denom/symbol.
func GetCurrency(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "currency [symbol]",
		Short: "get currency by chainID and denom/symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get", queryRoute), nil)
			if err != nil {
				return err
			}

			var out types.Currency
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
