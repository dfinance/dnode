// Message handler for currencies module.
package currencies

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/currencies/msgs"
)

// Handler for currencies messages, provess issue/destroy messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {

		case msgs.MsgDestroyCurrency:
			return handleMsgDestroy(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle destroy message.
func handleMsgDestroy(ctx sdk.Context, keeper Keeper, msg msgs.MsgDestroyCurrency) sdk.Result {
	err := keeper.DestroyCurrency(ctx, msg.ChainID, msg.Symbol, msg.Recipient, msg.Amount, msg.Spender)

	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
