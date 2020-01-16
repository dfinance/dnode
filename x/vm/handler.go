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
			return keeper.DeployContract(ctx, msg).Result()

		default:
			errMsg := fmt.Sprintf("unrecognized vm msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
