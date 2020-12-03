// +build unit

package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Check call creation and verify all related resources created / updated.
func TestMSKeeper_SubmitCall(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	msg := NewMockMsMsg(true)

	// do same check for multiple submits (checking nextID key and queue work)
	for i := 0; i < 3; i++ {
		callID := dnTypes.NewIDFromUint64(uint64(i))
		callUniqueID := fmt.Sprintf("uniqueID_%d", i)
		creator := sdk.AccAddress(fmt.Sprintf("addr_%d", i))

		t.Logf("CallID: %s", callID)

		// check call doesn't exist yet
		{
			require.False(t, keeper.HasCall(ctx, callID))
			require.False(t, keeper.HasCallByUniqueID(ctx, callUniqueID))

			_, getErr := keeper.GetCall(ctx, callID)
			require.Error(t, getErr)

			_, getIDErr := keeper.GetCallIDByUniqueID(ctx, callUniqueID)
			require.Error(t, getIDErr)
		}

		// submit (create)
		{
			err := keeper.SubmitCall(ctx, msg, callUniqueID, creator)
			require.NoError(t, err)
		}

		// get callID and check nextCallID is incremented
		{
			getterCallID := keeper.GetLastCallID(ctx)
			require.True(t, callID.Equal(getterCallID))
			require.True(t, keeper.nextCallID(ctx).Equal(callID.Incr()))
		}

		// check Has-methods
		{
			require.True(t, keeper.HasCall(ctx, callID))
			require.True(t, keeper.HasCallByUniqueID(ctx, callUniqueID))
		}

		// check added to the queue
		{
			store := ctx.KVStore(keeper.storeKey)
			require.True(t, store.Has(types.GetQueueKey(callID, 0)))

			iterator := keeper.GetQueueIteratorStartEnd(ctx, 0, 1)
			queueLen := 0
			for ; iterator.Valid(); iterator.Next() {
				queueLen++
			}
			iterator.Close()
			require.Equal(t, i+1, queueLen)
		}

		// check Get ID by uniqueID
		{
			getterCallID, err := keeper.GetCallIDByUniqueID(ctx, callUniqueID)
			require.NoError(t, err)
			require.True(t, callID.Equal(getterCallID))
		}

		// check Get
		{
			call, err := keeper.GetCall(ctx, callID)
			require.NoError(t, err)

			require.True(t, call.ID.Equal(callID))
			require.Equal(t, callUniqueID, call.UniqueID)
			require.Equal(t, creator.String(), call.Creator.String())
			require.False(t, call.Approved)
			require.False(t, call.Rejected)
			require.False(t, call.Executed)
			require.False(t, call.Failed)
			require.Empty(t, call.Error)
			require.Equal(t, ctx.BlockHeight(), call.Height)
			require.NotNil(t, call.Msg)
			require.Equal(t, msg.Type(), call.Msg.Type())
			require.Equal(t, msg.Route(), call.Msg.Route())
			require.NoError(t, call.CanBeVoted())
		}

		// check confirmed by creator
		{
			votes, err := keeper.GetVotes(ctx, callID)
			require.NoError(t, err)
			require.Len(t, votes, 1)
			require.Equal(t, creator.String(), votes[0].String())
		}

		// check call Update
		{
			inCall, err := keeper.GetCall(ctx, callID)
			require.NoError(t, err)

			inCall.Approved = true
			keeper.StoreCall(ctx, inCall)

			outCall, err := keeper.GetCall(ctx, callID)
			require.NoError(t, err)
			require.Error(t, outCall.CanBeVoted())
		}
	}
}

// Test SubmitCall with invalid inputs including unregistered routing and failing handler.
func TestMSKeeper_SubmitCall_InvalidInputs(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	sender := sdk.AccAddress("addr1")
	uniqueID1, uniqueID2 := "uniqueID1", "uniqueID2"

	// create call
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, uniqueID1, sender))
	}

	// invalid: existing uniqueID
	{
		msg := NewMockMsMsg(true)
		require.Error(t, keeper.SubmitCall(ctx, msg, uniqueID1, sender))
	}

	// invalid: call
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, uniqueID1, sdk.AccAddress{}))
	}

	// invalid: non-existing msg route
	{
		msg := CustomMockMsMsg{msType: MockMsgType, msRoute: "invalid"}
		require.Error(t, keeper.SubmitCall(ctx, msg, uniqueID2, sdk.AccAddress{}))
	}

	// invalid: failing dry-run
	{
		msg := CustomMockMsMsg{msType: MockMsgType, msRoute: MockMsgRouteErr}
		require.Error(t, keeper.SubmitCall(ctx, msg, uniqueID2, sdk.AccAddress{}))
	}
}
