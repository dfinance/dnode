package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/currencies/client/cli"
	types2 "github.com/dfinance/dnode/x/currencies/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types2.ModuleName,
		Short: "Querying commands for the currencies module",
	}

	queryCmd.AddCommand(
		sdkClient.GetCommands(
			cli.GetIssue(types2.ModuleName, cdc),
			cli.GetCurrency(types2.ModuleName, cdc),
			cli.GetWithdraw(types2.ModuleName, cdc),
			cli.GetWithdraws(types2.ModuleName, cdc),
		)...)

	return queryCmd
}

// GetTxCmd returns module tx commands.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types2.ModuleName,
		Short: "Currency transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostMsIssueCurrency(cdc),
		cli.PostWithdrawCurrency(cdc),
	)...)

	return txCmd
}
