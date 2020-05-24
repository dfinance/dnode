package client

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/currencies_register/client/cli"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// GetQueryCmd returns module query commands.
func GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the orders module",
	}

	queryCmd.AddCommand(sdkClient.GetCommands(
		cli.GetCmdInfo(types.RouterKey, cdc),
	)...)

	return queryCmd
}
