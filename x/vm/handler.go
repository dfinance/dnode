package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// New message handler for PoA module.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		// settings actual context for ds.
		// TODO: move it to base app and set before transaction execution maybe? or find way to have actual context always
		keeper.SetDSContext(ctx)

		switch msg := msg.(type) {
		case MsgDeployModule:
			return handleMsgDeploy(ctx, keeper, msg)

		case MsgExecuteScript:
			return handleMsgScript(ctx, keeper, msg)

		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized vm msg type: %v", msg.Type())
		}
	}
}

func handleMsgScript(ctx sdk.Context, keeper Keeper, msg MsgExecuteScript) (*sdk.Result, error) {
	if err := keeper.ExecuteScript(ctx, msg); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgDeploy(ctx sdk.Context, keeper Keeper, msg MsgDeployModule) (*sdk.Result, error) {
	if err := keeper.DeployContract(ctx, msg); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
