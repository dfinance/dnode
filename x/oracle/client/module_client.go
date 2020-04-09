package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/oracle/client/cli"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// Returns get commands for this module.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the oracle module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetCmdCurrentPrice(types.ModuleName, cdc),
		cli.GetCmdRawPrices(types.ModuleName, cdc),
		cli.GetCmdAssets(types.ModuleName, cdc),
		cli.GetCmdAssetCodeHex(),
	)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
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
