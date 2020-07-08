// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Check queue iterators and queue calls add / remove.
func TestMSKeeper_Queue(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	callIDs0 := make([]dnTypes.ID, 0)
	callIDs1 := make([]dnTypes.ID, 0)
	for i := 0; i < 3; i++ {
		callIDs0 = append(callIDs0, dnTypes.NewIDFromUint64(uint64(i)))
		callIDs1 = append(callIDs1, dnTypes.NewIDFromUint64(uint64(i+3)))
	}

	getIteratedCalls := func(iterator sdk.Iterator) []dnTypes.ID {
		ids := make([]dnTypes.ID, 0)
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var id dnTypes.ID
			cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &id)
			ids = append(ids, id)
		}
		return ids
	}

	checkIds := func(a, b []dnTypes.ID) {
		require.Equal(t, len(a), len(b))

		aUint := make([]uint64, 0, len(a))
		bUint := make([]uint64, 0, len(a))
		for i := 0; i < len(a); i++ {
			aUint = append(aUint, a[i].UInt64())
			bUint = append(bUint, b[i].UInt64())
		}

		require.EqualValues(t, aUint, bUint)
	}

	// remove non-existing
	{
		keeper.RemoveCallFromQueue(ctx, callIDs0[0], 0)
	}

	// add all calls for blockHeight 0 and 1
	{
		for _, id := range callIDs0 {
			keeper.addCallToQueue(ctx, id, 0)
		}
		for _, id := range callIDs1 {
			keeper.addCallToQueue(ctx, id, 1)
		}
	}

	// check iterator over all heights
	{
		inIds := append(callIDs0, callIDs1...)
		outIds := getIteratedCalls(keeper.GetQueueIteratorStartEnd(ctx, 0, 1))
		checkIds(inIds, outIds)
	}

	// check iterator for height 1 only (StartEnd)
	{
		outIds := getIteratedCalls(keeper.GetQueueIteratorStartEnd(ctx, 1, 1))
		checkIds(callIDs1, outIds)
	}

	// check iterator for height 0 only (Till)
	{
		outIds := getIteratedCalls(keeper.GetQueueIteratorTill(ctx, 0))
		checkIds(callIDs0, outIds)
	}

	// remove one call for each height
	{
		keeper.RemoveCallFromQueue(ctx, callIDs0[1], 0)
		callIDs0 = append(callIDs0[:1], callIDs0[2:]...)

		keeper.RemoveCallFromQueue(ctx, callIDs1[0], 1)
		callIDs1 = callIDs1[1:]

		inIds := append(callIDs0, callIDs1...)
		outIds := getIteratedCalls(keeper.GetQueueIteratorTill(ctx, 1))
		checkIds(inIds, outIds)
	}
}
