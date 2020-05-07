package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/order/internal/types"
)

func (k Keeper) LockOrderCoins(ctx sdk.Context, order types.Order) error {
	coin, err := order.QuoteAssetCoin()
	if err != nil {
		return sdkErrors.Wrapf(err, "creating lock coin for %s order", order.Direction.String())
	}

	if err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, order.Owner, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "SendCoinsFromAccountToModule for %s order: %v", order.Direction.String(), err)
	}

	return nil
}

func (k Keeper) UnlockOrderCoins(ctx sdk.Context, order types.Order) error {
	coin, err := order.QuoteAssetCoin()
	if err != nil {
		return sdkErrors.Wrapf(err, "creating unlock coin for %s order", order.Direction.String())
	}

	if err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, order.Owner, sdk.NewCoins(coin)); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "SendCoinsFromModuleToAccount for %s order: %v", order.Direction.String(), err)
	}

	return nil
}
