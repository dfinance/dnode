package orderbook

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders"
)

// EndBlocker iterates over Orders module orders, processes them and returns back to the Order module.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	iterator := k.GetOrderIterator(ctx)
	defer iterator.Close()

	matcherPool := NewMatcherPool(k.GetLogger(ctx))
	for ; iterator.Valid(); iterator.Next() {
		order := orders.Order{}
		ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)

		if err := matcherPool.AddOrder(order); err != nil {
			panic(err)
		}
	}

	resultCnt := 0
	for _, result := range matcherPool.Process() {
		k.ProcessOrderFills(ctx, result.OrderFills)
		k.SetHistoryItem(ctx, NewHistoryItem(ctx, result))

		resultCnt++
		ctx.EventManager().EmitEvent(NewClearanceEvent(result))
	}

	if resultCnt > 0 {
		ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))
	}

	return []abci.ValidatorUpdate{}
}
