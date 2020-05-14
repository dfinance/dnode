package orders

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/orders/internal/types"
)

// EndBlocker iterates over active orders and cancels them by TTL timeout condition.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	now := ctx.BlockTime()
	iterator := k.GetIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		order := types.Order{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)

		if now.Sub(order.CreatedAt) >= order.Ttl {
			k.GetLogger(ctx).Info(fmt.Sprintf("order canceled by TTL: %s", order.ID.String()))
			k.RevokeOrder(ctx, order.ID)
		}
	}

	return []abci.ValidatorUpdate{}
}
