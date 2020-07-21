package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/vm/client/cli"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the VM module (including compiler)",
	}

	compileCommands := sdkClient.GetCommands(
		cli.Compile(cdc),
	)
	for _, cmd := range compileCommands {
		cmd.Flags().String(vm_client.FlagCompilerAddr, config.DefaultCompilerAddr, vm_client.FlagCompilerUsage)
		cmd.Flags().String(vm_client.FlagOutput, "", "--to-file ./compiled.mv")
	}

	commands := sdkClient.GetCommands(
		cli.GetData(types.ModuleName, cdc),
		cli.GetTxVMStatus(cdc),
	)
	commands = append(commands, compileCommands...)

	queryCmd.AddCommand(commands...)

	return queryCmd
}

// GetTxCmd returns module tx commands.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "VM transactions subcommands",
	}

	compileCommands := sdkClient.PostCommands(
		cli.ExecuteScript(cdc),
	)
	for _, cmd := range compileCommands {
		cmd.Flags().String(vm_client.FlagCompilerAddr, config.DefaultCompilerAddr, vm_client.FlagCompilerUsage)
		txCmd.AddCommand(cmd)
	}

	commands := sdkClient.PostCommands(
		cli.DeployContract(cdc),
		sdkClient.LineBreak,
		cli.UpdateStdlibProposal(cdc),
	)
	commands = append(commands, compileCommands...)

	txCmd.AddCommand(commands...)

	return txCmd
}
