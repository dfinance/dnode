// Get cli commands.
package client

import (
	"github.com/WingsDao/wings-blockchain/x/vm/client/cli"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
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
