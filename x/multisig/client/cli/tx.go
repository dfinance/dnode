package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	msMsg "wings-blockchain/x/multisig/msgs"
	"strconv"
)
func PostConfirmCall(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "confirm-call [callId]",
		Short: "confirm call by multisig",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			callId, err := strconv.ParseUint(args[0], 10, 8)

			if err  != nil {
				return err
			}

			msg := msMsg.NewMsgConfirmCall(callId, cliCtx.GetFromAddress())

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func PostRevokeConfirm(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-confirm [callId]",
		Short: "revoke confirmation from call by id",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			callId, err := strconv.ParseUint(args[0], 10, 8)

			if err  != nil {
				return err
			}

			msg := msMsg.NewMsgRevokeConfirm(callId, cliCtx.GetFromAddress())

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}