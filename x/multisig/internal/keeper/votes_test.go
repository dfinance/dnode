// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Check call confirmation.
func TestMSKeeper_ConfirmCall(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	msg := NewMockMsMsg(true)
	callID := dnTypes.NewIDFromUint64(0)
	addr1, addr2 := sdk.AccAddress("addr1"), sdk.AccAddress("addr2")

	// fail: non-existing callID
	{
		require.Error(t, keeper.ConfirmCall(ctx, callID, addr1))
	}

	// create call
	{
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr1))
	}

	// ok: confirm call from different account
	{
		votesBefore, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votesBefore, 1)

		require.NoError(t, keeper.ConfirmCall(ctx, callID, addr2))

		votesAfter, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votesAfter, 2)
		require.Equal(t, addr1.String(), votesAfter[0].String())
		require.Equal(t, addr2.String(), votesAfter[1].String())
	}

	// fail: confirm with already existing vote
	{
		require.Error(t, keeper.ConfirmCall(ctx, callID, addr2))
	}
}

// Check call confirmation for already approved call.
func TestMSKeeper_ConfirmCall_Approved(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	callID := dnTypes.NewIDFromUint64(0)
	addr := sdk.AccAddress("addr1")

	// create call
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr))
	}

	// update call (change state)
	{
		call, err := keeper.GetCall(ctx, callID)
		require.NoError(t, err)

		call.Approved = true
		keeper.StoreCall(ctx, call)
	}

	// fail: call not accepts votes
	{
		require.Error(t, keeper.ConfirmCall(ctx, callID, addr))
	}
}

// Check revoking call confirmations.
func TestMSKeeper_RevokeConfirmation(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	callID := dnTypes.NewIDFromUint64(0)
	addr1, addr2 := sdk.AccAddress("addr1"), sdk.AccAddress("addr2")

	// fail: non-existing callID
	{
		require.Error(t, keeper.RevokeConfirmation(ctx, callID, addr1))
	}

	// create call
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr1))
	}

	// fail: non-existing vote
	{
		require.Error(t, keeper.RevokeConfirmation(ctx, callID, addr2))
	}

	// add one more vote
	{
		require.NoError(t, keeper.ConfirmCall(ctx, callID, addr2))

		votes, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votes, 2)
	}

	// ok (one left)
	{
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr1))

		votes, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votes, 1)

		// check call is in thequeue
		{
			storage := ctx.KVStore(keeper.storeKey)
			require.True(t, storage.Has(types.GetVotesKey(callID)))
		}
	}

	// ok (no left)
	{
		votesBefore, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votesBefore, 1)

		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr2))

		votesAfter, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votesAfter, 0)

		// check call removed from the queue
		{
			storage := ctx.KVStore(keeper.storeKey)
			require.False(t, storage.Has(types.GetVotesKey(callID)))
		}
	}

	// fail: revoke already revoked vote
	{
		require.Error(t, keeper.RevokeConfirmation(ctx, callID, addr2))
	}
}

// Check revoking call confirmations for already rejected call (not accepting votes).
func TestMSKeeper_RevokeConfirmation_Rejected(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	callID := dnTypes.NewIDFromUint64(0)
	addr := sdk.AccAddress("addr1")

	// create call
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr))
	}

	// update call (change state)
	{
		call, err := keeper.GetCall(ctx, callID)
		require.NoError(t, err)

		call.Rejected = true
		keeper.StoreCall(ctx, call)
	}

	// fail: call not accepts votes
	{
		require.Error(t, keeper.RevokeConfirmation(ctx, callID, addr))
	}
}

// Check call votes counter.
func TestMSKeeper_GetConfirmationsCount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	callID := dnTypes.NewIDFromUint64(0)
	addr1, addr2 := sdk.AccAddress("addr1"), sdk.AccAddress("addr2")

	// fail: non-existing call
	{
		_, err := keeper.GetConfirmationsCount(ctx, callID)
		require.Error(t, err)
	}

	// create call, check count == 1
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr1))

		cnt, err := keeper.GetConfirmationsCount(ctx, callID)
		require.NoError(t, err)
		require.EqualValues(t, cnt, 1)
	}

	// confirm call, check count == 2
	{
		require.NoError(t, keeper.ConfirmCall(ctx, callID, addr2))

		cnt, err := keeper.GetConfirmationsCount(ctx, callID)
		require.NoError(t, err)
		require.EqualValues(t, cnt, 2)
	}
}

// Check votes getters.
func TestMSKeeper_HasVoteGetVotes(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	callID := dnTypes.NewIDFromUint64(0)
	addr1, addr2 := sdk.AccAddress("addr1"), sdk.AccAddress("addr2")

	// HasVote fail: non-existing call
	{
		_, err := keeper.HasVote(ctx, callID, addr1)
		require.Error(t, err)
	}

	// GetVotes fail: non-existing call
	{
		_, err := keeper.GetVotes(ctx, callID)
		require.Error(t, err)
	}

	// create call with one vote
	{
		msg := NewMockMsMsg(true)
		require.NoError(t, keeper.SubmitCall(ctx, msg, "uniqueID", addr1))
	}

	// HasVote ok
	{
		ok, err := keeper.HasVote(ctx, callID, addr1)
		require.NoError(t, err)
		require.True(t, ok)
	}

	// HasVote fail: call exists, but vote doesn't
	{
		ok, err := keeper.HasVote(ctx, callID, addr2)
		require.NoError(t, err)
		require.False(t, ok)
	}

	// add a vote
	{
		require.NoError(t, keeper.ConfirmCall(ctx, callID, addr2))
	}

	// GetVotes ok
	{
		votes, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votes, 2)
		require.Equal(t, addr1.String(), votes[0].String())
		require.Equal(t, addr2.String(), votes[1].String())
	}

	// remove all confirmation
	{
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr1))
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr2))
	}

	// GetVotes ok (no votes)
	{
		votes, err := keeper.GetVotes(ctx, callID)
		require.NoError(t, err)
		require.Len(t, votes, 0)
	}
}
