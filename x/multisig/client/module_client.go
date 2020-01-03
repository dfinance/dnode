package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	cli "wings-blockchain/x/multisig/client/cli"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	multisigQueryCmd := &cobra.Command{
		Use:   "multisig",
		Short: "Multisig commands for the validators module",
	}

	multisigQueryCmd.AddCommand(client.GetCommands(
		cli.GetLastId("multisig", cdc),
		cli.GetCall("multisig", cdc),
		cli.GetCalls("multisig", cdc),
		cli.GetCallByUniqueID("multisig", cdc),
	)...)

	return multisigQueryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	multisigTxCmd := &cobra.Command{
		Use:   "multisig",
		Short: "Multisig transactions subcommands",
	}

	multisigTxCmd.AddCommand(client.PostCommands(
		cli.PostConfirmCall(cdc),
		cli.PostRevokeConfirm(cdc),
	)...)

	return multisigTxCmd
}
