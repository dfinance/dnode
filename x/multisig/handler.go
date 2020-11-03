package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/keeper"
	"github.com/dfinance/dnode/x/multisig/internal/types"
	"github.com/dfinance/dnode/x/poa"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(msKeeper keeper.Keeper, poaKeeper poa.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSubmitCall:
			return handleMsgSubmitCall(ctx, msKeeper, poaKeeper, msg)
		case types.MsgConfirmCall:
			return handleMsgConfirmCall(ctx, msKeeper, poaKeeper, msg)
		case types.MsgRevokeConfirm:
			return handleMsgRevokeConfirm(ctx, msKeeper, poaKeeper, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized currencies msg type: %v", msg.Type())
		}
	}
}

// handleMsgSubmitCall handles MsgSubmitCall message.
func handleMsgSubmitCall(ctx sdk.Context, msKeeper keeper.Keeper, poaKeeper poa.Keeper, msg types.MsgSubmitCall) (*sdk.Result, error) {
	if !poaKeeper.HasValidator(ctx, msg.Creator) {
		return nil, sdkErrors.Wrap(types.ErrPoaNotValidator, msg.Creator.String())
	}

	if err := msKeeper.SubmitCall(ctx, msg.Msg, msg.UniqueID, msg.Creator); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgConfirmCall handles MsgConfirmCall message.
func handleMsgConfirmCall(ctx sdk.Context, msKeeper keeper.Keeper, poaKeeper poa.Keeper, msg types.MsgConfirmCall) (*sdk.Result, error) {
	if !poaKeeper.HasValidator(ctx, msg.Sender) {
		return nil, sdkErrors.Wrap(types.ErrPoaNotValidator, msg.Sender.String())
	}

	if err := msKeeper.ConfirmCall(ctx, msg.CallID, msg.Sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgRevokeConfirm handles MsgRevokeConfirm message.
func handleMsgRevokeConfirm(ctx sdk.Context, msKeeper keeper.Keeper, poaKeeper poa.Keeper, msg types.MsgRevokeConfirm) (*sdk.Result, error) {
	if err := msKeeper.RevokeConfirmation(ctx, msg.CallID, msg.Sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
