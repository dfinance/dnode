// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Check runtime genesis export.
func TestCurrenciesKeeper_ExportGenesis(t *testing.T) {
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
