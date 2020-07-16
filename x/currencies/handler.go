package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgWithdrawCurrency:
			return handleMsgWithdraw(ctx, keeper, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized currencies msg type: %v", msg.Type())
		}
	}
}

// handleMsgWithdraw handles MsgWithdrawCurrency message.
func handleMsgWithdraw(ctx sdk.Context, keeper keeper.Keeper, msg MsgWithdrawCurrency) (*sdk.Result, error) {
	if err := keeper.WithdrawCurrency(ctx, msg.Coin, msg.Spender, msg.PegZoneRecipient, msg.PegZoneChainID); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
