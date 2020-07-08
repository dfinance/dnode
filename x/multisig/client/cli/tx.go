package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// PostConfirmCall returns tx command which confirms an existing call.
func PostConfirmCall(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-call [callID]",
		Short: "Confirm call by multi signature",
		Example: "confirm-call 100",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			callID, err := helpers.ParseDnIDParam("callID", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgConfirmCall(callID, fromAddr)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"callID [uint]",
	})

	return cmd
}

// PostRevokeConfirm returns tx command which revokes an existing call's vote.
func PostRevokeConfirm(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-confirm [callID]",
		Short: "Revoke confirmation for call by ID",
		Example: "revoke-confirm 100",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			callID, err := helpers.ParseDnIDParam("callID", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgRevokeConfirm(callID, fromAddr)

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"callID [uint]",
	})

	return cmd
}
