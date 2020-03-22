package vm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// New message handler for PoA module.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// settings actual context for ds.
		// TODO: move it to base app and set before transaction execution maybe? or find way to have actual context always
		keeper.SetDSContext(ctx)

		switch msg := msg.(type) {
		case MsgDeployModule:
			return handleMsgDeploy(ctx, keeper, msg)

		case MsgExecuteScript:
			return handleMsgScript(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized vm msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgScript(ctx sdk.Context, keeper Keeper, msg MsgExecuteScript) sdk.Result {
	if err := keeper.ExecuteScript(ctx, msg); err != nil {
		return err.Result()
	}

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgDeploy(ctx sdk.Context, keeper Keeper, msg MsgDeployModule) sdk.Result {
	if err := keeper.DeployContract(ctx, msg); err != nil {
		return err.Result()
	}

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
