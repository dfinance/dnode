// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

// Test keeper set/get params.
func TestCCSKeeper_Params(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	inParams := types.CurrenciesParams{
		"a": types.CurrencyParams{
			Decimals:       1,
			BalancePathHex: "01",
			InfoPathHex:    "0A",
		},
		"b": types.CurrencyParams{
			Decimals:       2,
			BalancePathHex: "02",
			InfoPathHex:    "0B",
		},
	}

	check := func(outParams types.CurrenciesParams) {
		for denom, in := range inParams {
			out, ok := outParams[denom]
			require.True(t, ok)
			require.Equal(t, in.Decimals, out.Decimals)
			require.Equal(t, in.BalancePathHex, out.BalancePathHex)
			require.Equal(t, in.InfoPathHex, out.InfoPathHex)
		}
	}

	// check set / get
	{
		keeper.setCurrenciesParams(ctx, inParams)
		check(keeper.GetCurrenciesParams(ctx))
	}
	
	// check update
	{
		newDenom := "c"
		newParams := types.CurrencyParams{
			Decimals:       3,
			BalancePathHex: "03",
			InfoPathHex:    "0C",
		}
		keeper.updateCurrenciesParams(ctx, newDenom, newParams)

		inParams[newDenom] = newParams
		check(keeper.GetCurrenciesParams(ctx))
	}
}
