package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	"wings-blockchain/x/poa/client/cli"
)

// Returns get commands for this module
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	poaQueryCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA commands for the validators module",
	}

	poaQueryCmd.AddCommand(client.GetCommands(
		cli.GetValidator("poa", cdc),
		cli.GetValidators("poa", cdc),
		cli.GetMinMax("poa", cdc),
	)...)

	return poaQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	poaTxCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA transactions subcommands",
	}

	poaTxCmd.AddCommand(client.PostCommands(
		cli.PostMsAddValidator(cdc),
		cli.PostMsRemoveValidator(cdc),
		cli.PostMsReplaceValidator(cdc),
	)...)

	return poaTxCmd
}
