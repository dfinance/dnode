// Multisignature message handler implementation.
package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	poa "wings-blockchain/x/poa"

	"fmt"
	"wings-blockchain/x/multisig/msgs"
	"wings-blockchain/x/multisig/types"
)

// Handle messages for multisig module.
func NewHandler(keeper Keeper, poaKeeper poa.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case msgs.MsgSubmitCall:
			return handleMsgSubmitCall(ctx, keeper, poaKeeper, msg)

		case msgs.MsgConfirmCall:
			return handleMsgConfirmCall(ctx, keeper, poaKeeper, msg)

		case msgs.MsgRevokeConfirm:
			return handleMsgRevokeConfirm(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized multisig msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle message (MsgSubmitCall) to submit new call.
func handleMsgSubmitCall(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper, msg msgs.MsgSubmitCall) sdk.Result {
	if !poaKeeper.HasValidator(ctx, msg.Sender) {
		return types.ErrNotValidator(msg.Sender.String()).Result()
	}

	err := keeper.SubmitCall(ctx, msg.Msg, msg.UniqueID, msg.Sender)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// Handle message (MsgConfirmCall) to confirm call.
func handleMsgConfirmCall(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper, msg msgs.MsgConfirmCall) sdk.Result {
	if !poaKeeper.HasValidator(ctx, msg.Sender) {
		return types.ErrNotValidator(msg.Sender.String()).Result()
	}

	has, err := keeper.HasVote(ctx, msg.MsgId, msg.Sender)

	if has {
		if err != nil {
			return err.Result()
		}

		return types.ErrCallAlreadyApproved(msg.MsgId, msg.Sender.String()).Result()
	}

	err = keeper.Confirm(ctx, msg.MsgId, msg.Sender)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}

// Handle message (MsgRevokeConfirm) to revoke call confirmation.
func handleMsgRevokeConfirm(ctx sdk.Context, keeper Keeper, msg msgs.MsgRevokeConfirm) sdk.Result {
	if has, err := keeper.HasVote(ctx, msg.MsgId, msg.Sender); err != nil {
		return err.Result()
	} else if !has {
		return types.ErrCallNotApproved(msg.MsgId, msg.Sender.String()).Result()
	}

	err := keeper.RevokeConfirmation(ctx, msg.MsgId, msg.Sender)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
