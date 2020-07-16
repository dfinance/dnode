package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/keeper"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgSubmitCall:
			return handleMsgSubmitCall(ctx, k, msg)
		case types.MsgConfirmCall:
			return handleMsgConfirmCall(ctx, k, msg)
		case types.MsgRevokeConfirm:
			return handleMsgRevokeConfirm(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized currencies msg type: %v", msg.Type())
		}
	}
}

// handleMsgSubmitCall handles MsgSubmitCall message.
func handleMsgSubmitCall(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitCall) (*sdk.Result, error) {
	if err := k.CheckAddressIsPoaValidator(ctx, msg.Creator); err != nil {
		return nil, err
	}

	if err := k.SubmitCall(ctx, msg.Msg, msg.UniqueID, msg.Creator); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgConfirmCall handles MsgConfirmCall message.
func handleMsgConfirmCall(ctx sdk.Context, k keeper.Keeper, msg types.MsgConfirmCall) (*sdk.Result, error) {
	if err := k.CheckAddressIsPoaValidator(ctx, msg.Sender); err != nil {
		return nil, err
	}

	if err := k.ConfirmCall(ctx, msg.CallID, msg.Sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgRevokeConfirm handles MsgRevokeConfirm message.
func handleMsgRevokeConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgRevokeConfirm) (*sdk.Result, error) {
	if err := k.RevokeConfirmation(ctx, msg.CallID, msg.Sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
