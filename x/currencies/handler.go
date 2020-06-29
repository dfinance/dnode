package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/internal/keeper"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgDestroyCurrency:
			return handleMsgDestroy(ctx, keeper, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized currencies msg type: %v", msg.Type())
		}
	}
}

// handleMsgDestroy handles MsgDestroyCurrency message.
func handleMsgDestroy(ctx sdk.Context, keeper keeper.Keeper, msg MsgDestroyCurrency) (*sdk.Result, error) {
	if err := keeper.DestroyCurrency(ctx, msg.Denom, msg.Amount, msg.Spender, msg.Recipient, msg.ChainID); err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}
