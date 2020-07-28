// +build unit

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getTestGenesisState(id uint64) GenesisState {
	item := NewMockHistoryItem(id)
	return GenesisState{
		HistoryItems: HistoryItems{item},
	}
}

func TestOrderBook_Genesis_Valid(t *testing.T) {
	//validateGenesis ok
	{
		item1 := NewMockHistoryItem(1)
		item2 := NewMockHistoryItem(2)
		item3 := NewMockHistoryItem(2)

		state := GenesisState{
			HistoryItems: HistoryItems{
				item1,
				item2,
			},
		}
		require.NoError(t, state.Validate(time.Now(), 1))
		require.False(t, state.IsEmpty())

		require.False(t, GenesisState{HistoryItems: HistoryItems{item1}}.Equal(GenesisState{HistoryItems: HistoryItems{item2}}))
		require.True(t, GenesisState{HistoryItems: HistoryItems{item3}}.Equal(GenesisState{HistoryItems: HistoryItems{item2}}))
	}

	// duplicated items
	{
		state := getTestGenesisState(1)
		state.HistoryItems = append(state.HistoryItems, state.HistoryItems[0])
		err := state.Validate(time.Now(), 1)

		require.Error(t, err)
		require.Contains(t, err.Error(), "ID")
		require.Contains(t, err.Error(), "duplicated")
	}

	// GT blockHeight
	{
		state := getTestGenesisState(1)
		state.HistoryItems[0].BlockHeight = 2
		err := state.Validate(time.Now(), 1)

		require.Error(t, err)
		require.Contains(t, err.Error(), "blockHeight")
		require.Contains(t, err.Error(), "GT")
	}

	// GT blockHeight
	{
		state := getTestGenesisState(1)
		state.HistoryItems[0].BlockHeight = 2
		err := state.Validate(time.Now(), 1)

		require.Error(t, err)
		require.Contains(t, err.Error(), "blockHeight")
		require.Contains(t, err.Error(), "GT")
		// blockHeight validation off
		{
			err := state.Validate(time.Time{}, -1)
			require.Nil(t, err)
		}
	}

	// GT timestamp
	{
		state := getTestGenesisState(1)
		state.HistoryItems[0].Timestamp = time.Now().Add(time.Hour).Unix()
		err := state.Validate(time.Now(), 1)

		require.Error(t, err)
		require.Contains(t, err.Error(), "timestamp")
		require.Contains(t, err.Error(), "after")
		// timestamp validation off
		{
			err := state.Validate(time.Time{}, -1)
			require.Nil(t, err)
		}
	}
}
