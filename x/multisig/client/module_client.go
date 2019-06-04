package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	cli "wings-blockchain/x/multisig/client/cli"
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
		Use:   "multisig",
		Short: "Multisig commands for the validators module",
	}

	return currenciesQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	currenciesTxCmd := &cobra.Command{
		Use:   "multisig",
		Short: "Multisig transactions subcommands",
	}

	currenciesTxCmd.AddCommand(client.PostCommands(
		cli.PostAddValidatorCall(mc.cdc),
		cli.PostConfirmCall(mc.cdc),
	)...)

	return currenciesTxCmd
}
