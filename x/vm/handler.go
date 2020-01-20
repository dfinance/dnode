package vm

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// New message handler for PoA module.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeployContract:
			return handleMsgDeploy(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized vm msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgDeploy(ctx sdk.Context, keeper Keeper, msg MsgDeployContract) sdk.Result {
	events, err := keeper.DeployContract(ctx, msg)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Events: events,
	}
}
