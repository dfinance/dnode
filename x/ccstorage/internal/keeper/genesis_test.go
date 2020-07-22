// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

// Check genesis currencies created and params updated.
func TestCCSKeeper_InitGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	defGenesis := types.DefaultGenesisState()

	currencies := keeper.GetCurrencies(ctx)
	require.Equal(t, len(defGenesis.CurrenciesParams), len(currencies))

	for _, genParams := range defGenesis.CurrenciesParams {
		require.True(t, keeper.HasCurrency(ctx, genParams.Denom))

		foundCnt := 0
		for _, curParams := range currencies.ToParams() {
			if curParams.Denom == genParams.Denom {
				require.Equal(t, genParams.Decimals, curParams.Decimals)
				require.Equal(t, genParams.BalancePathHex, curParams.BalancePathHex)
				require.Equal(t, genParams.InfoPathHex, curParams.InfoPathHex)

				foundCnt++
			}
		}
		require.Equal(t, 1, foundCnt)
	}
}

// Check runtime genesis export.
func TestCCSKeeper_ExportGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	state := types.GenesisState{}
	bz := keeper.ExportGenesis(ctx)
	input.cdc.MustUnmarshalJSON(bz, &state)

	for _, curParams := range keeper.GetCurrencies(ctx).ToParams() {
		foundCnt, foundIdx := 0, 0
		for i, expParams := range state.CurrenciesParams {
			if curParams.Denom == expParams.Denom {
				require.Equal(t, expParams.Decimals, curParams.Decimals)
				require.Equal(t, expParams.BalancePathHex, curParams.BalancePathHex)
				require.Equal(t, expParams.InfoPathHex, curParams.InfoPathHex)

				foundCnt++
				foundIdx = i
			}
		}
		require.Equal(t, 1, foundCnt)
		state.CurrenciesParams = append(state.CurrenciesParams[:foundIdx], state.CurrenciesParams[foundIdx+1:]...)
	}
	require.Empty(t, state.CurrenciesParams)
}
