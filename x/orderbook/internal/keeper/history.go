package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

// HasHistoryItem checks if historyItem object with marketID and blockHeight exists.
func (k Keeper) HasHistoryItem(ctx sdk.Context, marketID dnTypes.ID, blockHeight int64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetHistoryItemKey(marketID, blockHeight))
}

// GetHistoryItem gets historyItem object by marketID and blockHeight.
func (k Keeper) GetHistoryItem(ctx sdk.Context, marketID dnTypes.ID, blockHeight int64) (types.HistoryItem, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetHistoryItemKey(marketID, blockHeight))
	if bz == nil {
		return types.HistoryItem{}, types.ErrWrongHistoryItem
	}

	item := types.HistoryItem{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &item); err != nil {
		panic(fmt.Errorf("historyItem unmarshal: %w", err))
	}

	return item, nil
}

// SetHistoryItem adds historyItem to the storage.
func (k Keeper) SetHistoryItem(ctx sdk.Context, item types.HistoryItem) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoryItemKey(item.MarketID, item.BlockHeight)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(item)
	store.Set(key, bz)
}

// GetHistoryItemsInBlockHeightRange return historyItems per marketID in blockHeight range.
func (k Keeper) GetHistoryItemsInBlockHeightRange(ctx sdk.Context, marketID dnTypes.ID, startHeight, endHeight int64) (types.HistoryItems, error) {
	store := ctx.KVStore(k.storeKey)
	startKey := types.GetHistoryItemKey(marketID, startHeight)
	endKey := types.GetHistoryItemKey(marketID, endHeight)

	iterator := store.Iterator(startKey, endKey)
	defer iterator.Close()

	items := types.HistoryItems{}
	for ; iterator.Valid(); iterator.Next() {
		item := types.HistoryItem{}
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &item); err != nil {
			return types.HistoryItems{}, sdkErrors.Wrap(types.ErrInternal, "historyItem unmarshal")
		}

		items = append(items, item)
	}

	return items, nil
}
