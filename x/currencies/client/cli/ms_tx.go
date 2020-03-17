// Multisignature currency module commands for CLI.
package cli

import (
	"fmt"
	"os"
	"strconv"

	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/currencies/msgs"
	msMsg "github.com/dfinance/dnode/x/multisig/msgs"
)

// Issue new currency command.
func PostMsIssueCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-issue-currency [symbol] [amount] [decimals] [recipient] [issueID]",
		Short: "issue new currency via multisignature",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			amount, isOk := sdk.NewIntFromString(args[1])
			if !isOk {
				return fmt.Errorf("%s argument %q is not a number, can't parse int", "amount", args[1])
			}

			decimals, err := strconv.ParseInt(args[2], 10, 8)
			if err != nil {
				return fmt.Errorf("%s argument %q is not a number, can't parse int", "decimals", args[2])
			}

			recipient, err := sdk.AccAddressFromBech32(args[3])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "recipient", args[3], err)
			}

			msgIssCurr := msgs.NewMsgIssueCurrency(args[0], amount, int8(decimals), recipient, args[4])
			msg := msMsg.NewMsgSubmitCall(msgIssCurr, args[4], cliCtx.GetFromAddress())

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
