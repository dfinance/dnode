package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		// setup current (actual) DS context (that is also done in the BeginBlock
		// TODO: move it to base app and set before transaction execution maybe? or find way to have actual context always
		k.SetDSContext(ctx)

		switch msg := msg.(type) {
		case MsgDeployModule:
			return handleMsgDeploy(ctx, k, msg)

		case MsgExecuteScript:
			return handleMsgScript(ctx, k, msg)

		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized vm msg type: %v", msg.Type())
		}
	}
}

// handleMsgScript handles MsgExecuteScript message.
func handleMsgScript(ctx sdk.Context, k Keeper, msg MsgExecuteScript) (*sdk.Result, error) {
	if err := k.ExecuteScript(ctx, msg); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgDeploy handles MsgDeployModule message.
func handleMsgDeploy(ctx sdk.Context, k Keeper, msg MsgDeployModule) (*sdk.Result, error) {
	if err := k.DeployContract(ctx, msg); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
