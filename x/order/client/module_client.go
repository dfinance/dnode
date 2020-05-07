package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/order/client/cli"
	"github.com/dfinance/dnode/x/order/internal/types"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the market module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetCmdListOrders(types.ModuleName, cdc),
	)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Market transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.GetCmdPostOrder(cdc),
		cli.GetCmdCancelOrder(cdc),
	)...,
	)

	return txCmd
}
