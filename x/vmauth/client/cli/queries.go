package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/helpers"
)

// GetAccountCmd returns a query cmd that return account state (same as std keeper query, but using VM balance resources).
func GetAccountCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			addr, err := helpers.ParseSdkAddressParam("address", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare request
			req := authTypes.QueryAccountParams{
				Address: addr,
			}

			bz, err := cdc.MarshalJSON(req)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", authTypes.QuerierRoute, authTypes.QueryAccount), bz)
			if err != nil {
				return err
			}

			var acc exported.Account
			if err := cdc.UnmarshalJSON(res, &acc); err != nil {
				return err
			}

			return cliCtx.PrintOutput(acc)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"account address",
	})

	return flags.GetCommands(cmd)[0]
}
