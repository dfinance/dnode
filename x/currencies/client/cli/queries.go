package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// GetCurrency returns query command that return currency by denom.
func GetCurrency(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "currency [denom]",
		Short:   "Get currency by denom",
		Example: "currency dfi",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// prepare request
			req := types.CurrencyReq{Denom: args[0]}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryCurrency), bz)
			if err != nil {
				return err
			}

			var out types.Currency
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetIssue returns query command that return issue by id.
func GetIssue(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "issue [issueID]",
		Short: "Get issue by ID",
		Example: "issue issue1",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// prepare request
			req := types.IssueReq{ID: args[0]}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryIssue), bz)
			if err != nil {
				return err
			}

			var out types.Issue
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}

// GetDestroys returns query command that lists all destroy objects with filters and pagination.
func GetDestroys(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destroys",
		Short:   "Get destroys list by page and limit",
		Example: "destroys --page=1 --limit=10",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			page, limit, err := helpers.ParsePaginationParams(viper.GetString(flags.FlagPage), viper.GetString(flags.FlagLimit), helpers.ParamTypeCliFlag)
			if err != nil {
				return err
			}

			// prepare request
			req := types.DestroysReq{
				Page:  page,
				Limit: limit,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDestroys), bz)
			if err != nil {
				return err
			}

			var out types.Destroys
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.AddPaginationCmdFlags(cmd)

	return cmd
}

// GetDestroy returns query command that return destroy by id.
func GetDestroy(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "destroy [destroyID]",
		Short:   "Get destroy by ID",
		Example: "destroy 0",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			id, err := helpers.ParseDnIDParam("destroyID", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := types.DestroyReq{
				ID: id,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDestroy), bz)
			if err != nil {
				return err
			}

			var out types.Destroy
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
