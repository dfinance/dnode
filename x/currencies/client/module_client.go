// Implements getters for query and transaction CLI commands.
package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/currencies/client/cli"
	"github.com/dfinance/dnode/x/currencies/types"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the currencies module",
	}

	queryCmd.AddCommand(
		sdkClient.GetCommands(
			cli.GetIssue("currencies", cdc),
			cli.GetCurrency("currencies", cdc),
			cli.GetDestroy("currencies", cdc),
			cli.GetDestroys("currencies", cdc),
		)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Currency transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostMsIssueCurrency(cdc),
		cli.PostDestroyCurrency(cdc),
	)...)

	return txCmd
}
