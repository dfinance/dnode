package cli

import (
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	"wings-blockchain/x/currencies/queries"
	"github.com/cosmos/cosmos-sdk/codec"
	"fmt"
)

func GetDestroys(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "destroys [page] [limit]",
		Args:  cobra.ExactArgs(2),
		Short: "get destroys list by limit and page",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroys/%s/%s", queryRoute, args[0], args[1]), nil)

			if err != nil {
				return err
			}

			var out queries.QueryDestroysRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetDestroy(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy [destroyID]",
		Short: "get destroy by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroy/%s", queryRoute, args[0]), nil)

			if err != nil {
				return err
			}

			var out queries.QueryDestroyRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get denoms list
func GetIssue(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "issue [issueID]",
		Short: "get issue by id",
		Args: cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			var out queries.QueryIssueRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get currency by denom/symbol
func GetCurrency(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "currency [symbol]",
		Short: "get currency by chainID and denom/symbol",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			var out queries.QueryCurrencyRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

