package cli

import (
	"fmt"
	"os"
	"strconv"

	"wings-blockchain/x/currencies/msgs"
	msMsg "wings-blockchain/x/multisig/msgs"

	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

// Issue new currency command
func PostMsIssueCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-issue-currency [currencyId] [symbol] [amount] [decimals] [recipient] [issueID] [uniqueID]",
		Short: "issue new currency via multisignature",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			amount, isOk := sdk.NewIntFromString(args[2])

			if !isOk {
				return fmt.Errorf("can't parse int %s as amount", args[2])
			}

			decimals, err := strconv.ParseInt(args[3], 10, 8)

			if err != nil {
				return err
			}

			recipient, err := sdk.AccAddressFromBech32(args[4])

			if err != nil {
				return err
			}

			currencyId, isOk := sdk.NewIntFromString(args[0])

			if !isOk {
				return fmt.Errorf("can't parse int %s as currency id", args[0])
			}

			msgIssCurr := msgs.NewMsgIssueCurrency(currencyId, args[1], amount, int8(decimals), recipient, args[5])
			msg := msMsg.NewMsgSubmitCall(msgIssCurr, args[5], cliCtx.GetFromAddress())

			err = msg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
