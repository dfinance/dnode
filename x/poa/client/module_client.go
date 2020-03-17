package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/poa/client/cli"
)

// Return query commands for PoA module.
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

// Returns transactions commands for this module.
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
