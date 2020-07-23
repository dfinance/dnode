// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test genesis params validation.
func TestCCS_GenesisParams_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		param := CurrencyParams{"dfi",0, "0102", "AABB"}
		require.NoError(t, param.Validate())
	}

	// fail: invalid denom
	{
		param1 := CurrencyParams{"dfi1",0, "0102", "AABB"}
		require.Error(t, param1.Validate())
	}

	// fail: empty path
	{
		param1 := CurrencyParams{"dfi",0, "", "AABB"}
		require.Error(t, param1.Validate())

		param2 := CurrencyParams{"dfi",0, "0102", ""}
		require.Error(t, param2.Validate())
	}

	// fail: invalid hex path
	{
		param1 := CurrencyParams{"dfi",0, "z", "AABB"}
		require.Error(t, param1.Validate())

		param2 := CurrencyParams{"dfi",0, "0102", "z"}
		require.Error(t, param2.Validate())
	}
}

// Test genesis validation.
func TestCCS_Genesis_Validate(t *testing.T) {
	t.Parallel()

	state := GenesisState{}

	// ok: empty
	{
		require.NoError(t, state.Validate())
	}

	// ok: new 1
	{
		state.CurrenciesParams = append(state.CurrenciesParams, CurrencyParams{
			Denom:          "dfi",
			Decimals:       0,
			BalancePathHex: "0102",
			InfoPathHex:    "0A0B",
		})
		require.NoError(t, state.Validate())
	}

	// ok: new 2
	{
		state.CurrenciesParams = append(state.CurrenciesParams, CurrencyParams{
			Denom:          "btc",
			Decimals:       8,
			BalancePathHex: "1112",
			InfoPathHex:    "1A1B",
		})
		require.NoError(t, state.Validate())
	}

	// fail: duplicate
	{
		state.CurrenciesParams = append(state.CurrenciesParams, CurrencyParams{
			Denom:          "btc",
			Decimals:       4,
			BalancePathHex: "3132",
			InfoPathHex:    "3A3B",
		})
		require.Error(t, state.Validate())
	}
}