package orderbook

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/orderbook/internal/keeper"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

// EndBlocker iterates over Orders module orders, processes them and returns back to the Order module.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	iterator := k.GetOrderIterator(ctx)
	defer iterator.Close()

	matcherPool := keeper.NewMatcherPool(k.GetLogger(ctx))
	for ; iterator.Valid(); iterator.Next() {
		order := orderTypes.Order{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)

		if err := matcherPool.AddOrder(order); err != nil {
			panic(err)
		}
	}

	for _, result := range matcherPool.Process() {
		k.ProcessOrderFills(ctx, result.OrderFills)
		k.SetHistoryItem(ctx, types.NewHistoryItem(ctx, result))
	}

	return []abci.ValidatorUpdate{}
}
