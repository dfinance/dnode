// Message handler for currencies module.
package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/msgs"
)

// Handler for currencies messages, provess issue/destroy messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {

		case msgs.MsgDestroyCurrency:
			return handleMsgDestroy(ctx, keeper, msg)

		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized currencies msg type: %v", msg.Type())
		}
	}
}

// Handle destroy message.
func handleMsgDestroy(ctx sdk.Context, keeper Keeper, msg msgs.MsgDestroyCurrency) (*sdk.Result, error) {
	if err := keeper.DestroyCurrency(ctx, msg.ChainID, msg.Symbol, msg.Recipient, msg.Amount, msg.Spender); err != nil {
		return nil, err
	}

	return &sdk.Result{}, nil
}
