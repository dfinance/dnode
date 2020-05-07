package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/order/internal/types"
)

func (k Keeper) Has(ctx sdk.Context, id dnTypes.ID) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetOrderKey(id))
}

func (k Keeper) Get(ctx sdk.Context, id dnTypes.ID) (types.Order, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOrderKey(id))
	if bz == nil {
		return types.Order{}, types.ErrWrongOrderID
	}

	order := types.Order{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &order); err != nil {
		panic(fmt.Errorf("order unmarshal: %w", err))
	}

	return order, nil
}

func (k Keeper) Set(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(order.ID)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(order)
	store.Set(key, bz)
}

func (k Keeper) Del(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(id)
	store.Delete(key)
}

func (k Keeper) GetIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)

	return sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
}

func (k Keeper) GetReverseIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)

	return sdk.KVStoreReversePrefixIterator(store, types.OrderKeyPrefix)
}
