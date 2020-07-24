// +build unit

package keeper

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

func checkCallsEqual(t *testing.T, callA, callB types.Call) {
	require.Equal(t, callA.ID.String(), callB.ID.String())
	require.Equal(t, callA.UniqueID, callB.UniqueID)
	require.True(t, callA.Creator.Equals(callB.Creator))
	require.Equal(t, callA.Approved, callB.Approved)
	require.Equal(t, callA.Executed, callB.Executed)
	require.Equal(t, callA.Failed, callB.Failed)
	require.Equal(t, callA.Rejected, callB.Rejected)
	require.Equal(t, callA.Error, callB.Error)
	require.Equal(t, callA.MsgRoute, callB.MsgRoute)
	require.Equal(t, callA.MsgType, callB.MsgType)
	require.Equal(t, callA.Height, callB.Height)

	msgA, errA := json.Marshal(callA.Msg)
	require.NoError(t, errA)
	msgB, errB := json.Marshal(callB.Msg)
	require.NoError(t, errB)
	require.Equal(t, msgA, msgB)
}

func checkVotesEqual(t *testing.T, votesA, votesB types.Votes) {
	require.EqualValues(t, len(votesA), len(votesB))
	for i := 0; i < len(votesA); i++ {
		require.EqualValues(t, votesA[i], votesB[i])
	}
}

// Check genesis import/export with various cases.
func TestMSKeeper_Genesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, inputCtx, cdc := input.target, input.ctx, input.cdc

	// check params
	{
		ctx, _ := inputCtx.CacheContext()
		state := types.GenesisState{
			Parameters: types.Params{
				IntervalToExecute: 123,
			},
		}

		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))
		require.EqualValues(t, 123, keeper.GetIntervalToExecute(ctx))
		require.Nil(t, keeper.getLastCallID(ctx))

		// export
		{
			var exportedState types.GenesisState
			cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

			require.Equal(t, state.Parameters.IntervalToExecute, exportedState.Parameters.IntervalToExecute)
		}
	}

	// check calls
	{
		ctx, _ := inputCtx.CacheContext()

		msg := NewMockMsMsg(true)
		addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

		state := types.DefaultGenesisState()
		state.CallItems = []types.GenesisCallItem{
			{
				Call: types.Call{
					ID:       dnTypes.NewIDFromUint64(0),
					UniqueID: "call_0",
					Creator:  addr1,
					Msg:      msg,
					MsgRoute: msg.Route(),
					MsgType:  msg.Type(),
				},
				Votes: types.Votes{},
			},
			{
				Call: types.Call{
					ID:       dnTypes.NewIDFromUint64(1),
					UniqueID: "call_1",
					Creator:  addr2,
					Approved: true,
					Executed: true,
					Msg:      msg,
					MsgRoute: msg.Route(),
					MsgType:  msg.Type(),
				},
				Votes: types.Votes{
					addr1,
					addr2,
				},
			},
		}
		lastId := dnTypes.NewIDFromUint64(1)
		state.LastCallID = &lastId

		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))

		require.NotNil(t, keeper.getLastCallID(ctx))
		require.Equal(t, state.LastCallID.String(), keeper.getLastCallID(ctx).String())

		for _, callItem := range state.CallItems {
			require.True(t, keeper.HasCall(ctx, callItem.Call.ID))
			require.True(t, keeper.HasCallByUniqueID(ctx, callItem.Call.UniqueID))

			votes, _ := keeper.GetVotes(ctx, callItem.Call.ID)
			require.EqualValues(t, callItem.Votes, votes)

			// check msg is routable
			call, err := keeper.GetCall(ctx, callItem.Call.ID)
			require.NoError(t, err)

			handler := keeper.GetRouteHandler(call.Msg.Route())
			require.NotNil(t, handler)
		}

		// export
		{
			var exportedState types.GenesisState
			cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

			require.Len(t, exportedState.CallItems, len(state.CallItems))
			for _, exportItem := range exportedState.CallItems {
				var initItem types.GenesisCallItem
				for _, callItem := range state.CallItems {
					if exportItem.Call.ID.Equal(callItem.Call.ID) {
						initItem = callItem
						break
					}
				}

				checkCallsEqual(t, initItem.Call, exportItem.Call)
				checkVotesEqual(t, initItem.Votes, exportItem.Votes)
			}
		}
	}

	// check queue
	{
		ctx, _ := inputCtx.CacheContext()

		msg := NewMockMsMsg(true)
		addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

		state := types.DefaultGenesisState()
		state.CallItems = []types.GenesisCallItem{
			{
				Call: types.Call{
					ID:       dnTypes.NewIDFromUint64(0),
					UniqueID: "call_0",
					Creator:  addr1,
					Msg:      msg,
					MsgRoute: msg.Route(),
					MsgType:  msg.Type(),
				},
				Votes: types.Votes{},
			},
			{
				Call: types.Call{
					ID:       dnTypes.NewIDFromUint64(1),
					UniqueID: "call_1",
					Creator:  addr1,
					Msg:      msg,
					MsgRoute: msg.Route(),
					MsgType:  msg.Type(),
				},
				Votes: types.Votes{},
			},
			{
				Call: types.Call{
					ID:       dnTypes.NewIDFromUint64(2),
					UniqueID: "call_2",
					Creator:  addr1,
					Msg:      msg,
					MsgRoute: msg.Route(),
					MsgType:  msg.Type(),
				},
				Votes: types.Votes{},
			},
		}
		state.QueueItems = []types.GenesisQueueItem{
			{
				CallID:      state.CallItems[2].Call.ID,
				BlockHeight: 0,
			},
			{
				CallID:      state.CallItems[1].Call.ID,
				BlockHeight: 0,
			},
		}
		lastId := dnTypes.NewIDFromUint64(2)
		state.LastCallID = &lastId

		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))

		require.NotNil(t, keeper.getLastCallID(ctx))
		require.Equal(t, state.LastCallID.String(), keeper.getLastCallID(ctx).String())

		queueIdx := 0
		iterator := keeper.GetQueueIteratorTill(ctx, 0)
		for ; iterator.Valid(); iterator.Next() {
			var callID dnTypes.ID
			cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &callID)
			_, blockHeith := types.MustParseQueueKey(iterator.Key())

			found := false
			for _, initItem := range state.QueueItems {
				if initItem.CallID.Equal(callID) && initItem.BlockHeight == blockHeith {
					found = true
					break
				}
			}
			require.True(t, found)
			queueIdx++
		}
		require.Equal(t, len(state.QueueItems), queueIdx)

		// export
		{
			var exportedState types.GenesisState
			cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

			require.Len(t, exportedState.QueueItems, len(state.QueueItems))
			for _, exportItem := range exportedState.QueueItems {
				found := false
				for _, initItem := range state.QueueItems {
					if exportItem.CallID.Equal(initItem.CallID) && exportItem.BlockHeight == initItem.BlockHeight {
						found = true
						break
					}
				}
				require.True(t, found)
			}
		}
	}
}
