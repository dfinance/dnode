package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// PostWithdrawCurrency returns tx command which post a new withdraw request.
func PostWithdrawCurrency(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-currency [denom] [amount] [pegZoneSpender] [pegZoneChainID]",
		Short:   "Withdraw issued currency, reducing spender balance",
		Example: "withdraw-currency dfi 100 {account} testnet --from {account}",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			amount, err := helpers.ParseSdkIntParam("amount", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgWithdrawCurrency(args[0], amount, fromAddr, args[2], args[3])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"currency denomination symbol",
		"reduce coin amount",
		"spender address for PegZone",
		"chainID for PegZone blockchain",
	})

	return cmd
}
