// Multisignature message handler implementation.
package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/multisig/msgs"
	"github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/poa"
)

// Handle messages for multisig module.
func NewHandler(keeper Keeper, poaKeeper poa.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case msgs.MsgSubmitCall:
			return handleMsgSubmitCall(ctx, keeper, poaKeeper, msg)

		case msgs.MsgConfirmCall:
			return handleMsgConfirmCall(ctx, keeper, poaKeeper, msg)

		case msgs.MsgRevokeConfirm:
			return handleMsgRevokeConfirm(ctx, keeper, msg)

		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized multisig msg type: %v", msg.Type())
		}
	}
}

// Handle message (MsgSubmitCall) to submit new call.
func handleMsgSubmitCall(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper, msg msgs.MsgSubmitCall) (*sdk.Result, error) {
	if !poaKeeper.HasValidator(ctx, msg.Sender) {
		return nil, sdkErrors.Wrap(types.ErrNotValidator, msg.Sender.String())
	}

	if err := keeper.SubmitCall(ctx, msg.Msg, msg.UniqueID, msg.Sender); err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}

// Handle message (MsgConfirmCall) to confirm call.
func handleMsgConfirmCall(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper, msg msgs.MsgConfirmCall) (*sdk.Result, error) {
	if !poaKeeper.HasValidator(ctx, msg.Sender) {
		return nil, sdkErrors.Wrap(types.ErrNotValidator, msg.Sender.String())
	}

	has, err := keeper.HasVote(ctx, msg.MsgId, msg.Sender)

	if has {
		if err != nil {
			return nil, err
		}
		return nil, sdkErrors.Wrapf(types.ErrCallAlreadyApproved, "%d by %s", msg.MsgId, msg.Sender.String())
	}

	if err := keeper.Confirm(ctx, msg.MsgId, msg.Sender); err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}

// Handle message (MsgRevokeConfirm) to revoke call confirmation.
func handleMsgRevokeConfirm(ctx sdk.Context, keeper Keeper, msg msgs.MsgRevokeConfirm) (*sdk.Result, error) {
	if has, err := keeper.HasVote(ctx, msg.MsgId, msg.Sender); err != nil {
		return nil, err
	} else if !has {
		return nil, sdkErrors.Wrapf(types.ErrCallNotApproved, "%d by %s", msg.MsgId, msg.Sender.String())
	}

	if err := keeper.RevokeConfirmation(ctx, msg.MsgId, msg.Sender); err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}
