// Keeper queue to manage calls.
package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/WingsDao/wings-blockchain/x/multisig/types"
)

// Adding a new call to queue.
func (keeper Keeper) addCallToQueue(ctx sdk.Context, callId uint64, height int64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.GetQueueKey(callId, height), keeper.cdc.MustMarshalBinaryLengthPrefixed(callId))
}

// Remove a call from queue.
func (keeper Keeper) removeCallFromQueue(ctx sdk.Context, callId uint64, height int64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(types.GetQueueKey(callId, height))
}

// Getting queue iterator from block height to end block height.
func (keeper Keeper) GetQueueIteratorStartEnd(ctx sdk.Context, startHeight, endHeight int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)

	return store.Iterator(types.GetPrefixQueue(startHeight), sdk.PrefixEndBytes(types.GetPrefixQueue(endHeight)))
}

// Get queue iterator till.
func (keeper Keeper) GetQueueIteratorTill(ctx sdk.Context, endHeight int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)

	return store.Iterator(types.PrefixQueue, sdk.PrefixEndBytes(types.GetPrefixQueue(endHeight)))
}
