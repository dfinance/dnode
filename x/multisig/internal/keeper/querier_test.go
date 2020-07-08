// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Check getCall query with various votes length.
func TestMSQuery_GetCall(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	checkCallResp := func(callID dnTypes.ID, uniqueID string, votes types.Votes) {
		req := types.CallReq{CallID: callID}
		res, err := queryGetCall(keeper, ctx, abci.RequestQuery{Data: cdc.MustMarshalJSON(req)})
		require.NoError(t, err)

		var resp types.CallResp
		cdc.MustUnmarshalJSON(res, &resp)

		// brief call check
		require.Equal(t, callID.UInt64(), resp.Call.ID.UInt64())
		require.Equal(t, uniqueID, resp.Call.UniqueID)

		// votes check
		require.Len(t, resp.Votes, len(votes))
		for i := 0; i < len(resp.Votes); i++ {
			require.Equal(t, votes[i].String(), resp.Votes[i].String())
		}
	}

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	callID := dnTypes.NewIDFromUint64(0)
	uniqueID := "uniqueID"

	// create call
	{
		msg := NewMockMsMsg(true)
		err := keeper.SubmitCall(ctx, msg, uniqueID, addr1)
		require.NoError(t, err)
	}

	// check getCall: one vote
	{
		checkCallResp(callID, uniqueID, types.Votes{addr1})
	}

	// check getCall: two votes
	{
		require.NoError(t, keeper.ConfirmCall(ctx, callID, addr2))
		checkCallResp(callID, uniqueID, types.Votes{addr1, addr2})
	}

	// check getCall: no votes
	{
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr1))
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID, addr2))
		checkCallResp(callID, uniqueID, types.Votes{})
	}
}

// Check getCalls query requesting active calls in the queue.
func TestMSQuery_GetCalls(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	// init genesis
	keeper.InitDefaultGenesis(ctx)

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	callID1, callID2, callID3 := dnTypes.NewIDFromUint64(0), dnTypes.NewIDFromUint64(1), dnTypes.NewIDFromUint64(2)
	uniqueID1, uniqueID2, uniqueID3 := "unique1", "unique2", "unique3"
	msg := NewMockMsMsg(true)

	// create 1st call with no votes (that one is removed from the queue)
	{
		require.NoError(t, keeper.SubmitCall(ctx, msg, uniqueID1, addr1))
		require.NoError(t, keeper.RevokeConfirmation(ctx, callID1, addr1))
	}

	// create 2nd call with one vote
	{
		require.NoError(t, keeper.SubmitCall(ctx, msg, uniqueID2, addr1))
	}

	// create 3rd call with two votes
	{
		require.NoError(t, keeper.SubmitCall(ctx, msg, uniqueID3, addr1))
		require.NoError(t, keeper.ConfirmCall(ctx, callID3, addr2))
	}

	// request
	var resp types.CallsResp
	{
		res, err := queryGetCalls(keeper, ctx)
		require.NoError(t, err)
		cdc.MustUnmarshalJSON(res, &resp)
	}

	// check
	{
		require.Len(t, resp, 2)

		require.Equal(t, resp[0].Call.ID.UInt64(), callID2.UInt64())
		require.Equal(t, resp[0].Call.UniqueID, uniqueID2)
		require.Len(t, resp[0].Votes, 1)
		require.Equal(t, addr1.String(), resp[1].Votes[0].String())

		require.Equal(t, resp[1].Call.ID.UInt64(), callID3.UInt64())
		require.Equal(t, resp[1].Call.UniqueID, uniqueID3)
		require.Len(t, resp[1].Votes, 2)
		require.Equal(t, addr1.String(), resp[1].Votes[0].String())
		require.Equal(t, addr2.String(), resp[1].Votes[1].String())
	}
}
