// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Test keeper set/get params.
func TestCurrenciesKeeper_Params(t *testing.T) {
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

	keeper.SetCurrenciesParams(ctx, inParams)
	outParams := keeper.GetCurrenciesParams(ctx)
	
	for denom, in := range inParams {
		out, ok := outParams[denom]
		require.True(t, ok)
		require.Equal(t, in.Decimals, out.Decimals)
		require.Equal(t, in.BalancePathHex, out.BalancePathHex)
		require.Equal(t, in.InfoPathHex, out.InfoPathHex)
	}
}
