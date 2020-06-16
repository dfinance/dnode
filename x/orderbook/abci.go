package orderbook

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	orderTypes "github.com/dfinance/dnode/x/orders"
)

// EndBlocker iterates over Orders module orders, processes them and returns back to the Order module.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	iterator := k.GetOrderIterator(ctx)
	defer iterator.Close()

	matcherPool := NewMatcherPool(k.GetLogger(ctx))
	for ; iterator.Valid(); iterator.Next() {
		order := orderTypes.Order{}
		ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)

		if err := matcherPool.AddOrder(order); err != nil {
			panic(err)
		}
	}

	for _, result := range matcherPool.Process() {
		k.ProcessOrderFills(ctx, result.OrderFills)
		k.SetHistoryItem(ctx, NewHistoryItem(ctx, result))

		ctx.EventManager().EmitEvent(NewClearanceEvent(result.MarketID, result.ClearanceState.Price))
	}

	return []abci.ValidatorUpdate{}
}
