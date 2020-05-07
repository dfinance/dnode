package order

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/order/internal/types"
)

func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	now := ctx.BlockTime()
	iterator := k.GetIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		order := types.Order{}
		if err := types.ModuleCdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &order); err != nil {
			panic(fmt.Errorf("order unmarshal: %w", err))
		}

		if now.Sub(order.CreatedAt) >= order.Ttl {
			k.GetLogger(ctx).Info(fmt.Sprintf("order canceled by TTL: %s", order.ID.String()))
			k.CancelOrder(ctx, order.ID)
		}
	}

	return []abci.ValidatorUpdate{}
}
