package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/poa/client/cli"
	types2 "github.com/dfinance/dnode/x/poa/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types2.ModuleName,
		Short: "Querying commands for the PoA module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetValidator(types2.ModuleName, cdc),
		cli.GetValidators(types2.ModuleName, cdc),
		cli.GetMinMax(types2.ModuleName, cdc),
	)...)

	return queryCmd
}

// GetTxCmd returns module tx commands.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   types2.ModuleName,
		Short: "PoA transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.PostMsAddValidator(cdc),
		cli.PostMsRemoveValidator(cdc),
		cli.PostMsReplaceValidator(cdc),
	)...)

	return txCmd
}
