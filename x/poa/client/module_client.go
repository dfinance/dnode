package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"wings-blockchain/x/poa/client/cli"
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
		Use:   "validators",
		Short: "Querying commands for the validators module",
	}

	return currenciesQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	currenciesTxCmd := &cobra.Command{
		Use:   "validators",
		Short: "Validators transactions subcommands",
	}

	currenciesTxCmd.AddCommand(client.PostCommands(
		cli.PostAddValidator(mc.cdc),
		cli.PostRemoveValidator(mc.cdc),
		cli.PostReplaceValidator(mc.cdc),
	)...)

	return currenciesTxCmd
}
