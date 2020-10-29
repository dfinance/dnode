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
		// add to the queue if call is not handled yet
		if err := call.CanBeVoted(); err == nil {
			k.RemoveCallFromQueue(ctx, call.ID, call.Height)
			k.addCallToQueue(ctx, call.ID, 0)
		}

		// update call
		call.Height = 0
		k.StoreCall(ctx, call)
	}

	return nil
}
