package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// GetCmdAddMarket returns tx command which adds a market object.
func GetCmdAddMarket(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [base_denom] [quote_denom]",
		Short:   "Add a new market",
		Example: "add dfi eth",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			baseDenom, quoteDenom := args[0], args[1]
			if err := helpers.ValidateDenomParam("base_denom", baseDenom, helpers.ParamTypeCliArg); err != nil {
				return err
			}
			if err := helpers.ValidateDenomParam("quote_denom", quoteDenom, helpers.ParamTypeCliArg); err != nil {
				return err
			}

			// message send
			msg := types.NewMsgCreateMarket(fromAddr, baseDenom, quoteDenom)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"base currency denomination symbol",
		"quote currency denomination symbol",
	})

	return cmd
}
