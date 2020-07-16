package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/helpers"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

// Has checks if order object with ID exists.
func (k Keeper) Has(ctx sdk.Context, id dnTypes.ID) bool {
	k.modulePerms.AutoCheck(types.PermReader)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetOrderKey(id))
}

// Get gets order object by ID.
func (k Keeper) Get(ctx sdk.Context, id dnTypes.ID) (types.Order, error) {
	k.modulePerms.AutoCheck(types.PermReader)

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

// GetList return all active orders.
func (k Keeper) GetList(ctx sdk.Context) (retOrders types.Orders, retErr error) {
	k.modulePerms.AutoCheck(types.PermReader)

	iterator := k.GetIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		order := types.Order{}
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &order); err != nil {
			retErr = fmt.Errorf("order unmarshal: %w", err)
			return
		}
		retOrders = append(retOrders, order)
	}

	return
}

// GetListFiltered returns order objects filtered by params.
func (k Keeper) GetListFiltered(ctx sdk.Context, params types.OrdersReq) (types.Orders, error) {
	k.modulePerms.AutoCheck(types.PermReader)

	orders, err := k.GetList(ctx)
	if err != nil {
		return types.Orders{}, err
	}

	paramsMarketID, _ := dnTypes.NewIDFromString(params.MarketID)
	filteredOrders := make(types.Orders, 0, len(orders))
	for _, o := range orders {
		match := true

		if params.OwnerFilter() && !o.Owner.Equals(params.Owner) {
			match = false
		}
		if params.MarketIDFilter() && !o.Market.ID.Equal(paramsMarketID) {
			match = false
		}
		if params.DirectionFilter() && !o.Direction.Equal(params.Direction) {
			match = false
		}

		if match {
			filteredOrders = append(filteredOrders, o)
		}
	}

	start, end, err := helpers.PaginateSlice(len(filteredOrders), params.Page, params.Limit)
	if err != nil {
		return types.Orders{}, err
	}

	return filteredOrders[start:end], nil
}

// GetIterator return order object iterator (direct sort order).
func (k Keeper) GetIterator(ctx sdk.Context) sdk.Iterator {
	k.modulePerms.AutoCheck(types.PermReader)

	store := ctx.KVStore(k.storeKey)

	return sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
}

// GetIterator return order object iterator (reverse sort order).
func (k Keeper) GetReverseIterator(ctx sdk.Context) sdk.Iterator {
	k.modulePerms.AutoCheck(types.PermReader)

	store := ctx.KVStore(k.storeKey)

	return sdk.KVStoreReversePrefixIterator(store, types.OrderKeyPrefix)
}


// set creates / overwrites order object.
func (k Keeper) set(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(order.ID)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(order)
	store.Set(key, bz)
}

// del removes order object.
func (k Keeper) del(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(id)
	store.Delete(key)
}
