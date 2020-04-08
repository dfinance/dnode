package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/poa/client/cli"
)

// Return query commands for PoA module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA commands for the validators module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetValidator("poa", cdc),
		cli.GetValidators("poa", cdc),
		cli.GetMinMax("poa", cdc),
	)...)

	return queryCmd
}

// Returns transactions commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "poa",
		Short: "PoA transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostMsAddValidator(cdc),
		cli.PostMsRemoveValidator(cdc),
		cli.PostMsReplaceValidator(cdc),
	)...)

	return txCmd
}
