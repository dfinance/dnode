package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/oracle/client/cli"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "oracle",
		Short: "Querying commands for the oracle module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetCmdCurrentPrice("oracle", cdc),
		cli.GetCmdRawPrices("oracle", cdc),
		cli.GetCmdAssets("oracle", cdc),
		cli.GetCmdAssetCodeHex(),
	)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "oracle",
		Short:                      "Oracle transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.GetCmdPostPrice(cdc),
		cli.GetCmdAddOracle(cdc),
		cli.GetCmdSetOracles(cdc),
		cli.GetCmdSetAsset(cdc),
		cli.GetCmdAddAsset(cdc),
	)...,
	)

	return txCmd
}
