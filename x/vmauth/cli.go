package vmauth

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
)

// GetAccountCmd returns a query account that will display the state of the
// account at a given address.
func GetAccountCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			key, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "address", args[0], err)
			}

			bz, err := cdc.MarshalJSON(types.QueryAccountParams{
				Address: key,
			})
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData("custom/acc/account", bz)

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

	return flags.GetCommands(cmd)[0]
}
