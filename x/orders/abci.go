package orders

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// EndBlocker iterates over active orders and cancels them by TTL timeout condition.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	now := ctx.BlockTime()
	prevEventsCnt := len(ctx.EventManager().Events())
	iterator := k.GetIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		order := Order{}
		ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &order)

		if now.Sub(order.CreatedAt) >= order.Ttl {
			k.GetLogger(ctx).Info(fmt.Sprintf("order canceled by TTL: %s", order.ID.String()))
			if err := k.RevokeOrder(ctx, order.ID); err != nil {
				k.GetLogger(ctx).Error(fmt.Sprintf("Revoking order %q by TTL: %v", order.ID, err))
			}
		}
	}

	if curEventsCnt := len(ctx.EventManager().Events()); curEventsCnt != prevEventsCnt {
		ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))
	}

	return []abci.ValidatorUpdate{}
}
