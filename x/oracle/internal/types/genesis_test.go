// +build unit

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getTestGenesisState() GenesisState {
	return GenesisState{
		Params:        DefaultParams(),
		CurrentPrices: CurrentPrices{NewMockCurrentPrice("btc_xfi", 10001, 1000)},
	}
}

func TestOracles_Genesis_Valid(t *testing.T) {
	//validateGenesis ok
	{
		order1 := NewMockCurrentPrice("btc_xfi", 10001, 1000)
		order2 := NewMockCurrentPrice("eth_xfi", 20002, 2000)
		order3 := NewMockCurrentPrice("xfi_btc", 30003, 3000)
		orderT := &order3
		order4 := *orderT

		state := getTestGenesisState()
		state.CurrentPrices = CurrentPrices{order1, order2, order3}
		require.NoError(t, state.Validate(time.Now().Add(time.Second)))
		require.False(t, state.IsEmpty())

		require.False(t, GenesisState{CurrentPrices: CurrentPrices{order2}}.Equal(GenesisState{CurrentPrices: CurrentPrices{order3}}))
		require.True(t, GenesisState{CurrentPrices: CurrentPrices{order3}}.Equal(GenesisState{CurrentPrices: CurrentPrices{order4}}))
	}

	// wrong id
	{
		state := getTestGenesisState()
		state.CurrentPrices = append(state.CurrentPrices, state.CurrentPrices[0])
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicated")
		require.Contains(t, err.Error(), "asset_code")
	}

	// wrong received_at
	{
		state := getTestGenesisState()
		state.CurrentPrices[0].ReceivedAt = time.Now().Add(time.Minute)
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "received_at")
		require.Contains(t, err.Error(), "after block time")
	}

	// wrong received_at, no validate
	{
		state := getTestGenesisState()
		err := state.Validate(time.Time{})

		require.Nil(t, err)
	}
}
