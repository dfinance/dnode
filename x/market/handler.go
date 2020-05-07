package market

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/market/internal/types"
)

// NewHandler handles all market type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgCreateMarket:
			return HandleMsgCreateMarket(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized market message type: %T", msg)
		}
	}
}

func HandleMsgCreateMarket(ctx sdk.Context, k Keeper, msg types.MsgCreateMarket) (*sdk.Result, error) {
	market, err := k.Add(ctx, msg.Nominee.String(), msg.BaseAssetDenom, msg.QuoteAssetDenom, msg.BaseAssetDecimals)
	if err != nil {
		return nil, err
	}

	res, err := ModuleCdc.MarshalBinaryLengthPrefixed(market)
	if err != nil {
		return nil, fmt.Errorf("result marshal: %w", err)
	}

	return &sdk.Result{
		Data:   res,
		Events: ctx.EventManager().Events(),
	}, nil
}
