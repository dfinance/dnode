package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	historyItems, err := k.GetHistoryItemsList(ctx)
	if err != nil {
		return fmt.Errorf("retrieving HistoryItem list: %w", err)
	}

	// remove all but the latest history item for each market
	historyItemsSet := make(map[string]types.HistoryItem)
	for _, curItem := range historyItems {
		existingItem, found := historyItemsSet[curItem.MarketID.String()]
		if !found || curItem.BlockHeight > existingItem.BlockHeight {
			historyItemsSet[curItem.MarketID.String()] = curItem
		}
		store.Delete(types.GetHistoryItemKey(curItem.MarketID, curItem.BlockHeight))
	}
	for _, item := range historyItemsSet {
		item.BlockHeight = 0
		k.SetHistoryItem(ctx, item)
	}

	return nil
}
