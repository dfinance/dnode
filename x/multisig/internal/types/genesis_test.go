// +build unit

package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func TestMS_Genesis_Validate(t *testing.T) {
	okCall := func(id uint64) Call {
		return Call{
			ID:       dnTypes.NewIDFromUint64(id),
			UniqueID: fmt.Sprintf("unique_%d", id),
			Creator:  sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
			Approved: true,
			Executed: true,
			Msg:      NewMockMsMsg("route", "type", true),
			MsgRoute: "route",
			MsgType:  "type",
			Height:   100,
		}
	}

	// fail: params
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute - 1,
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid call item
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  Call{ID: dnTypes.ID{}},
					Votes: Votes{},
				},
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid queue item: id
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(0),
					Votes: Votes{},
				},
			},
			QueueItems: []GenesisQueueItem{
				{
					CallID:      dnTypes.ID{},
					BlockHeight: 0,
				},
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid queue item: negative itemBlockHeight
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(0),
					Votes: Votes{},
				},
			},
			QueueItems: []GenesisQueueItem{
				{
					CallID:      dnTypes.NewZeroID(),
					BlockHeight: -1,
				},
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid queue item: itemBlockHeight > curBlockHeight
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(0),
					Votes: Votes{},
				},
			},
			QueueItems: []GenesisQueueItem{
				{
					CallID:      dnTypes.NewZeroID(),
					BlockHeight: 200,
				},
			},
		}
		require.Error(t, state.Validate(100))
	}
	// fail: invalid queue item: non-existing id
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(1),
					Votes: Votes{},
				},
			},
			QueueItems: []GenesisQueueItem{
				{
					CallID:      dnTypes.NewZeroID(),
					BlockHeight: 0,
				},
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid lastCallID: nil with existing calls
	{
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(1),
					Votes: Votes{},
				},
			},
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid lastCallID: not nil without existing calls
	{
		lastID := dnTypes.NewZeroID()
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			LastCallID: &lastID,
		}
		require.Error(t, state.Validate(-1))
	}
	// fail: invalid lastCallID: mismatch
	{
		lastID := dnTypes.NewZeroID()
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(0),
					Votes: Votes{},
				},
				{
					Call:  okCall(1),
					Votes: Votes{},
				},
			},
			LastCallID: &lastID,
		}
		require.Error(t, state.Validate(-1))
	}
	// ok
	{
		lastID := dnTypes.NewIDFromUint64(1)
		state := GenesisState{
			Parameters: Params{
				IntervalToExecute: MinIntervalToExecute,
			},
			CallItems: []GenesisCallItem{
				{
					Call:  okCall(0),
					Votes: Votes{},
				},
				{
					Call:  okCall(1),
					Votes: Votes{},
				},
			},
			LastCallID: &lastID,
		}
		require.NoError(t, state.Validate(-1))
	}
}
