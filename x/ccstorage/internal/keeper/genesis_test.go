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
	params := keeper.getCurrenciesParams(ctx)
	require.Equal(t, len(defGenesis.CurrenciesParams), len(params))

	for denom, genParams := range defGenesis.CurrenciesParams {
		require.True(t, keeper.HasCurrency(ctx, denom))

		paramParam, ok := params[denom]
		require.True(t, ok)

		require.Equal(t, genParams.Decimals, paramParam.Decimals)
		require.Equal(t, genParams.BalancePathHex, paramParam.BalancePathHex)
		require.Equal(t, genParams.InfoPathHex, paramParam.InfoPathHex)
	}
}

// Check runtime genesis export.
func TestCCSKeeper_ExportGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	defaultGenesis := types.DefaultGenesisState()
	ctx, keeper := input.ctx, input.keeper

	state := types.GenesisState{}
	bz := keeper.ExportGenesis(ctx)
	input.cdc.MustUnmarshalJSON(bz, &state)

	for defDenom, defParams := range defaultGenesis.CurrenciesParams {
		for rtDenom, rtParams := range state.CurrenciesParams {
			if defDenom == rtDenom {
				require.Equal(t, defParams.Decimals, rtParams.Decimals)
				require.Equal(t, defParams.BalancePathHex, rtParams.BalancePathHex)
				require.Equal(t, defParams.InfoPathHex, rtParams.InfoPathHex)

				delete(state.CurrenciesParams, rtDenom)
				break
			}
		}
	}
	require.Empty(t, state.CurrenciesParams)
}
