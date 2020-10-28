package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context) error {
	// reset call objects and calls queue entries
	// queue modifications resets call confirmation timeout
	calls := k.getCalls(ctx)
	for _, call := range calls {
		// queue
		k.RemoveCallFromQueue(ctx, call.ID, call.Height)
		k.addCallToQueue(ctx, call.ID, 0)
		// call
		call.Height = 0
		k.StoreCall(ctx, call)
	}

	return nil
}
