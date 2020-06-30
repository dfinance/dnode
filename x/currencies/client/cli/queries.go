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
	cmd := &cobra.Command{
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
	helpers.BuildCmdHelp(cmd, []string{
		"currency denomination symbol",
	})

	return cmd
}

// GetIssue returns query command that return issue by id.
func GetIssue(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
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
	helpers.BuildCmdHelp(cmd, []string{
		"unique issue ID",
	})

	return cmd
}

// GetWithdraws returns query command that lists all withdraw objects with filters and pagination.
func GetWithdraws(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraws",
		Short:   "Get withdraw list by page and limit",
		Example: "withdraws --page=1 --limit=10",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			page, limit, err := helpers.ParsePaginationParams(viper.GetString(flags.FlagPage), viper.GetString(flags.FlagLimit), helpers.ParamTypeCliFlag)
			if err != nil {
				return err
			}

			// prepare request
			req := types.WithdrawsReq{
				Page:  page,
				Limit: limit,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryWithdraws), bz)
			if err != nil {
				return err
			}

			var out types.Withdraws
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.AddPaginationCmdFlags(cmd)

	return cmd
}

// GetWithdraw returns query command that return withdraw by id.
func GetWithdraw(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw [withdrawID]",
		Short:   "Get withdraw by ID",
		Example: "withdraw 0",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			id, err := helpers.ParseDnIDParam("withdrawID", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := types.WithdrawReq{
				ID: id,
			}

			bz, err := cliCtx.Codec.MarshalJSON(req)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryWithdraw), bz)
			if err != nil {
				return err
			}

			var out types.Withdraw
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"withdraw unique ID",
	})

	return cmd
}
