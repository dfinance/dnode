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
		CurrentPrices: CurrentPrices{NewMockCurrentPrice("btc_dfi", 10000)},
	}
}

func TestOrders_Genesis_Valid(t *testing.T) {
	//validateGenesis ok
	{
		order1 := NewMockCurrentPrice("btc_dfi", 10000)
		order2 := NewMockCurrentPrice("eth_dfi", 20000)
		order3 := NewMockCurrentPrice("dfi_btc", 30000)
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
		require.Contains(t, err.Error(), "future date")
	}

	// wrong received_at, no validate
	{
		state := getTestGenesisState()
		err := state.Validate(time.Time{})

		require.Nil(t, err)
	}
}
