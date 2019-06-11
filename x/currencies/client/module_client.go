package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
	"wings-blockchain/x/currencies/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

// Creating new cli module
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// Returns get commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	currenciesQueryCmd := &cobra.Command{
		Use:   "currencies",
		Short: "Querying commands for the currencies module",
	}

	currenciesQueryCmd.AddCommand(
		client.GetCommands(
			cli.GetDenoms("currencies", mc.cdc),
			cli.GetCurrency("currencies", mc.cdc),
		)...)

	return currenciesQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	currenciesTxCmd := &cobra.Command{
		Use:   "currencies",
		Short: "Currency transactions subcommands",
	}

	currenciesTxCmd.AddCommand(client.PostCommands(
		cli.PostMsIssueCurrency(mc.cdc),
		cli.PostDestroyCurrency(mc.cdc),
	)...)

	return currenciesTxCmd
}
