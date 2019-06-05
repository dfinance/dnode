package cli


import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"strconv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/currencies/msgs"
	msMsg "wings-blockchain/x/multisig/msgs"
)

// Issue new currency command
func PostMsIssueCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-issue-currency [symbol] [amount] [decimals]",
		Short: "issue new currency via multisignature",
		Args:  cobra.ExactArgs(3),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)

			if err != nil {
				return err
			}

			decimals, err := strconv.ParseInt(args[2], 10, 8)

			if err != nil {
				return err
			}

			msgIssCurr := msgs.NewMsgIssueCurrency(args[0], amount, int8(decimals), cliCtx.GetFromAddress())
			msg := msMsg.NewMsgSubmitCall(msgIssCurr, cliCtx.GetFromAddress())

			err = msg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// Destroy currency
func PostMsDestroyCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: 	"ms-destroy-currency [symbol] [amount]",
		Short:  "destory issued currency via multisignature",
		Args: 	cobra.ExactArgs(2),
		RunE:   func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)

			if err != nil {
				return err
			}

			msgDesCur := msgs.NewMsgDestroyCurrency(args[0], amount, cliCtx.GetFromAddress())
			msg   	  := msMsg.NewMsgSubmitCall(msgDesCur, cliCtx.GetFromAddress())

			err = msg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
