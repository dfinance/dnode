package orders

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates order type messages handler.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgPostOrder:
			return handleMsgPostOrder(ctx, k, msg)
		case MsgRevokeOrder:
			return handleMsgCancelOrder(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized orders message type: %T", msg)
		}
	}
}

// handleMsgPostOrder handles MsgPostOrder message which creates a new order.
func handleMsgPostOrder(ctx sdk.Context, k Keeper, msg MsgPostOrder) (*sdk.Result, error) {
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

// handleMsgCancelOrder handles MsgRevokeOrder message which deletes a new order.
func handleMsgCancelOrder(ctx sdk.Context, k Keeper, msg MsgRevokeOrder) (*sdk.Result, error) {
	order, err := k.Get(ctx, msg.OrderID)
	if err != nil {
		return nil, err
	}

	if !order.Owner.Equals(msg.Owner) {
		return nil, sdkErrors.Wrap(ErrWrongOwner, "order owner mismatch")
	}

	if err := k.RevokeOrder(ctx, msg.OrderID); err != nil {
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
