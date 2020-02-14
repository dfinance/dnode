// Get cli commands.
package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"wings-blockchain/x/vm/client/cli"
	"wings-blockchain/x/vm/internal/types"
)

// Return TX commands for CLI.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM transactions commands",
	}

	txCmd.AddCommand(client.PostCommands(
		cli.DeployContract(cdc),
		cli.ExecuteScript(cdc),
	)...)

	return txCmd
}
