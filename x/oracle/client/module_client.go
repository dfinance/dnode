package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	cmd "wings-blockchain/x/oracle/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

// NewModuleClient creates client for the module
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group nameservice queries under a subcommand
	queryCmd := &cobra.Command{
		Use:   "oracle",
		Short: "Querying commands for the oracle module",
	}

	queryCmd.AddCommand(client.GetCommands(
		cmd.GetCmdCurrentPrice(mc.storeKey, mc.cdc),
		cmd.GetCmdRawPrices(mc.storeKey, mc.cdc),
		cmd.GetCmdAssets(mc.storeKey, mc.cdc),
	)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "oracle",
		Short: "Oracle transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		cmd.GetCmdPostPrice(mc.cdc),
	)...)

	return txCmd
}
