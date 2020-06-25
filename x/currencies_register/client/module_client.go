package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/currencies_register/client/cli"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "currencies_register transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		cli.AddCurrencyProposal(cdc),
	)...,
	)

	return txCmd
}

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the currencies_register module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetCmdInfo(types.RouterKey, cdc),
	)...)

	return queryCmd
}
