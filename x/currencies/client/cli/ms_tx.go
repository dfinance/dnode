package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
	msMsg "github.com/dfinance/dnode/x/multisig/msgs"
)

// PostMsIssueCurrency returns tx command which post a new multisig issue request.
func PostMsIssueCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-issue-currency [issueID] [denom] [amount] [decimals] [payee]",
		Short: "issue new currency via multisignature",
		Example: "ms-issue-currency issue1 dfi 100 18 {account} --from {account}",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			amount, err := helpers.ParseSdkIntParam("amount", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			decimals, err := helpers.ParseUint8Param("decimals", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			payee, err := helpers.ParseSdkAddressParam("payee", args[4], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send multisig message
			msg := types.NewMsgIssueCurrency(args[0], args[1], amount, decimals, payee)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			callMsg := msMsg.NewMsgSubmitCall(msg, args[0], fromAddr)
			if err := callMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{callMsg})
		},
	}
}
