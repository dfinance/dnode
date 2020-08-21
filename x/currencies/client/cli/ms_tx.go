package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
	msClient "github.com/dfinance/dnode/x/multisig/client"
)

// PostMsIssueCurrency returns tx command which post a new multisig issue request.
func PostMsIssueCurrency(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ms-issue [issueID] [coin] [payee]",
		Short:   "Issue new currency via multi signature, increasing payee coin balance",
		Example: "ms-issue issue1 100xfi {account} --from {account}",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			coin, err := helpers.ParseCoinParam("coin", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			payee, err := helpers.ParseSdkAddressParam("payee", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send multisig message
			msg := types.NewMsgIssueCurrency(args[0], coin, payee)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			callMsg := msClient.NewMsgSubmitCall(msg, args[0], fromAddr)
			if err := callMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{callMsg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"unique multi signature call ID",
		"currency denomination symbol and amount in Coin format (1.0 btc with 8 decimals -> 100000000btc)",
		"payee address (whose balance is increased)",
	})

	return cmd
}
