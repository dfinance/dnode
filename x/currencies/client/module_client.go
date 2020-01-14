// Implements getters for query and transaction CLI commands.
package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"wings-blockchain/x/currencies/client/cli"
	"wings-blockchain/x/currencies/types"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	currenciesQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the currencies module",
	}

	currenciesQueryCmd.AddCommand(
		client.GetCommands(
			cli.GetIssue("currencies", cdc),
			cli.GetCurrency("currencies", cdc),
			cli.GetDestroy("currencies", cdc),
			cli.GetDestroys("currencies", cdc),
		)...)

	return currenciesQueryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	currenciesTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Currency transactions subcommands",
	}

	currenciesTxCmd.AddCommand(client.PostCommands(
		cli.PostMsIssueCurrency(cdc),
		cli.PostDestroyCurrency(cdc),
	)...)

	return currenciesTxCmd
}
