package order

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/order/internal/types"
)

// NewHandler handles all order type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case types.MsgPostOrder:
			return HandleMsgPostOrder(ctx, k, msg)
		case types.MsgCancelOrder:
			return HandleMsgCancelOrder(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized order message type: %T", msg)
		}
	}
}

func HandleMsgPostOrder(ctx sdk.Context, k Keeper, msg types.MsgPostOrder) (*sdk.Result, error) {
	order, err := k.PostOrder(ctx, msg.Owner, msg.MarketID, msg.Direction, msg.Price, msg.Quantity, msg.TtlInSec)
	if err != nil {
		return nil, err
	}

	res, err := ModuleCdc.MarshalBinaryLengthPrefixed(order)
	if err != nil {
		return nil, fmt.Errorf("result marshal: %w", err)
	}

	return &sdk.Result{
		Data:   res,
		Events: ctx.EventManager().Events(),
	}, nil
}

func HandleMsgCancelOrder(ctx sdk.Context, k Keeper, msg types.MsgCancelOrder) (*sdk.Result, error) {
	order, err := k.Get(ctx, msg.OrderID)
	if err != nil {
		return nil, err
	}

	if !order.Owner.Equals(msg.Owner) {
		return nil, sdkErrors.Wrap(types.ErrWrongOwner, "order owner mismatch")
	}

	if err := k.CancelOrder(ctx, msg.OrderID); err != nil {
		return nil, err
	}

	res, err := ModuleCdc.MarshalBinaryLengthPrefixed(order)
	if err != nil {
		return nil, fmt.Errorf("result marshal: %w", err)
	}

	return &sdk.Result{
		Data:   res,
		Events: ctx.EventManager().Events(),
	}, nil
}
