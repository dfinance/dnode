package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/currencies/client/cli"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the currencies module",
	}

	queryCmd.AddCommand(
		sdkClient.GetCommands(
			cli.GetIssue(types.ModuleName, cdc),
			cli.GetCurrency(types.ModuleName, cdc),
			cli.GetCurrencies(types.ModuleName, cdc),
			cli.GetWithdraw(types.ModuleName, cdc),
			cli.GetWithdraws(types.ModuleName, cdc),
		)...)

	return queryCmd
}

// GetTxCmd returns module tx commands.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Currency transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostMsIssueCurrency(cdc),
		cli.PostMsUnstakeCurrency(cdc),
		cli.PostWithdrawCurrency(cdc),
		cli.AddCurrencyProposal(cdc),
	)...)

	return txCmd
}
