package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// RemoveCallFromQueue removes call from the queue.
func (k Keeper) RemoveCallFromQueue(ctx sdk.Context, id dnTypes.ID, height int64) {
	k.modulePerms.AutoCheck(types.PermWrite)

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetQueueKey(id, height))

	ctx.EventManager().EmitEvent(types.NewCallRemovedEvent(id))
}

// GetQueueIteratorStartEnd returns queue iterator within [start:end] blockHeight range.
func (k Keeper) GetQueueIteratorStartEnd(ctx sdk.Context, startHeight, endHeight int64) sdk.Iterator {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Iterator(types.GetPrefixQueueKey(startHeight), sdk.PrefixEndBytes(types.GetPrefixQueueKey(endHeight)))
}

// GetQueueIteratorTill returns queue iterator within [:end] blockHeight range.
// Get queue iterator till.
func (k Keeper) GetQueueIteratorTill(ctx sdk.Context, endHeight int64) sdk.Iterator {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Iterator(types.QueuePrefix, sdk.PrefixEndBytes(types.GetPrefixQueueKey(endHeight)))
}

// addCallToQueue add a new call to the queue.
func (k Keeper) addCallToQueue(ctx sdk.Context, id dnTypes.ID, height int64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetQueueKey(id, height), k.cdc.MustMarshalBinaryLengthPrefixed(id))
}
