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
	poaQueryCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA commands for the validators module",
	}

	poaQueryCmd.AddCommand(client.GetCommands(
		cli.GetValidator("poa", mc.cdc),
		cli.GetValidators("poa", mc.cdc),
		cli.GetMinMax("poa", mc.cdc),
	)...)

	return poaQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	poaTxCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA transactions subcommands",
	}

	poaTxCmd.AddCommand(client.PostCommands(
		cli.PostMsAddValidator(mc.cdc),
		cli.PostMsRemoveValidator(mc.cdc),
		cli.PostMsReplaceValidator(mc.cdc),
	)...)

	return poaTxCmd
}
