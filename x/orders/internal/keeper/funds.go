package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

// LockOrderCoins locks account funds defined by order on order posting.
// Coins transfer from Account to Module.
func (k Keeper) LockOrderCoins(ctx sdk.Context, order types.Order) error {
	coin, err := order.LockCoin()
	if err != nil {
		return sdkErrors.Wrap(err, "creating lock coin")
	}

	if err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, order.Owner, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "locking coins: %v", err)
	}

	return nil
}

// LockOrderCoins locks account funds defined by order on order canceling.
// Coins transfer from Module to Account.
func (k Keeper) UnlockOrderCoins(ctx sdk.Context, order types.Order) error {
	coin, err := order.LockCoin()
	if err != nil {
		return sdkErrors.Wrap(err, "creating unlock coin")
	}

	if err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, order.Owner, sdk.NewCoins(coin)); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "unlocking coins: %v", err)
	}

	return nil
}

// ExecuteOrderFills processes orderFills transfers fund on full / partial order execution.
// Refunding is done for bid order if clearancePrice is less that order target price.
// Order is removed from the store on full order fill.
// Order stays active on partial order fill (order quantity is reduced).
func (k Keeper) ExecuteOrderFills(ctx sdk.Context, orderFills types.OrderFills) {
	for _, orderFill := range orderFills {
		fillCoin, err := orderFill.FillCoin()
		if err != nil {
			k.GetLogger(ctx).Debug(orderFill.String())
			k.GetLogger(ctx).Error(fmt.Sprintf("creating fill coin: %v", err))
			continue
		}
		if _, err = k.bankKeeper.AddCoins(ctx, orderFill.Order.Owner, sdk.NewCoins(fillCoin)); err != nil {
			k.GetLogger(ctx).Debug(orderFill.String())
			panic(fmt.Sprintf("transfering fill coins: %v", err))
		}

		doRefund, refundCoin, err := orderFill.RefundCoin()
		if err != nil {
			k.GetLogger(ctx).Debug(orderFill.String())
			k.GetLogger(ctx).Error(fmt.Sprintf("creating refund coin: %v", err))
			continue
		}
		if doRefund {
			if refundCoin != nil {
				if _, err = k.bankKeeper.AddCoins(ctx, orderFill.Order.Owner, sdk.NewCoins(*refundCoin)); err != nil {
					k.GetLogger(ctx).Debug(orderFill.String())
					panic(fmt.Sprintf("adding refund coins: %v", err))
				}
			} else {
				k.GetLogger(ctx).Debug(orderFill.String())
				k.GetLogger(ctx).Info(fmt.Sprintf("order refund amount is too small: %s", orderFill.Order.ID))
			}
		}

		eventManager := ctx.EventManager()
		if orderFill.QuantityUnfilled.IsZero() {
			k.GetLogger(ctx).Info(fmt.Sprintf("order completely filled: %s", orderFill.Order.ID))
			k.Del(ctx, orderFill.Order.ID)
			eventManager.EmitEvent(types.NewFullyFilledOrderEvent(orderFill.Order))
		} else {
			k.GetLogger(ctx).Info(fmt.Sprintf("order partially filled: %s", orderFill.Order.ID))
			orderFill.Order.Quantity = orderFill.QuantityUnfilled
			orderFill.Order.UpdatedAt = ctx.BlockTime()
			k.Set(ctx, orderFill.Order)
			eventManager.EmitEvent(types.NewPartiallyFilledOrderEvent(orderFill.Order))
		}
	}

	if len(orderFills) > 0 {
		ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(types.ModuleName))
	}
}
