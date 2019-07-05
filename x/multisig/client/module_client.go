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
	multisigQueryCmd := &cobra.Command{
		Use:   "multisig",
		Short: "Multisig commands for the validators module",
	}

	multisigQueryCmd.AddCommand(client.GetCommands(
		cli.GetLastId("multisig", mc.cdc),
		cli.GetCall("multisig", mc.cdc),
		cli.GetCalls("multisig", mc.cdc),
	)...)

	return multisigQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	multisigTxCmd := &cobra.Command{
		Use:   "multisig",
		Short: "Multisig transactions subcommands",
	}

	multisigTxCmd.AddCommand(client.PostCommands(
		cli.PostConfirmCall(mc.cdc),
		cli.PostRevokeConfirm(mc.cdc),
	)...)

	return multisigTxCmd
}
