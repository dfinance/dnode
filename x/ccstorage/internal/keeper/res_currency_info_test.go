// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// Test keeper GetResStdCurrencyInfo method.
func TestCCSKeeper_GetStandardCurrencyInfo(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	defaultGenesis := types.DefaultGenesisState()
	ctx, keeper := input.ctx, input.keeper

	// ok
	{
		for _, params := range defaultGenesis.CurrenciesParams {
			curInfo, err := keeper.GetResStdCurrencyInfo(ctx, params.Denom)
			require.NoError(t, err)

			require.EqualValues(t, params.Denom, curInfo.Denom)
			require.EqualValues(t, params.Decimals, curInfo.Decimals)
			require.EqualValues(t, common_vm.StdLibAddress, curInfo.Owner)
			require.EqualValues(t, 0, curInfo.TotalSupply.Uint64())
			require.False(t, curInfo.IsToken)

			curBalancePath, err := keeper.GetCurrencyBalancePath(ctx, params.Denom)
			require.NoError(t, err)
			curInfoPath, err := keeper.GetCurrencyInfoPath(ctx, params.Denom)
			require.NoError(t, err)

			currency, err := keeper.GetCurrency(ctx, params.Denom)
			require.NoError(t, err)

			require.EqualValues(t, currency.BalancePath(), curBalancePath)
			require.EqualValues(t, currency.InfoPath(), curInfoPath)
		}
	}

	// fail: non-existing denom
	{
		_, err := keeper.GetResStdCurrencyInfo(ctx, "nonexisting")
		require.Error(t, err)
	}
}
