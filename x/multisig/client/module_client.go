package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/multisig/client/cli"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the multisig module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetLastId(types.ModuleName, cdc),
		cli.GetCall(types.ModuleName, cdc),
		cli.GetCalls(types.ModuleName, cdc),
		cli.GetCallByUniqueID(types.ModuleName, cdc),
	)...)

	return queryCmd
}

// GetTxCmd returns module tx commands.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Multisig transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostConfirmCall(cdc),
		cli.PostRevokeConfirm(cdc),
	)...)

	return txCmd
}
