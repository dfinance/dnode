package currencies

import (
	"github.com/cosmos/cosmos-sdk/types"
	"fmt"
	"wings-blockchain/x/currencies/msgs"
)

// Handler for currencies messages, provess issue/destory messages
func NewHandler(keeper Keeper) types.Handler {
	return func (ctx types.Context, msg types.Msg) types.Result {
		switch msg := msg.(type) {
		case msgs.MsgIssueCurrency:
			return handleMsgIssueCurrency(ctx, keeper, msg)

		case msgs.MsgDestroyCurrency:
			return handleMsgDestroy(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return types.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle issue message
func handleMsgIssueCurrency(ctx types.Context, keeper Keeper, msg msgs.MsgIssueCurrency) types.Result {
	err := keeper.IssueCurrency(ctx, msg.Symbol, msg.Amount, msg.Decimals, msg.Creator)

	if err != nil {
		return err.Result()
	}

	return types.Result{}
}

// Handle destory message
func handleMsgDestroy(ctx types.Context, keeper Keeper, msg msgs.MsgDestroyCurrency) types.Result {
	err := keeper.DestroyCurrency(ctx, msg.Symbol, msg.Amount, msg.Sender)

	if err != nil {
		return err.Result()
	}

	return types.Result{}
}

