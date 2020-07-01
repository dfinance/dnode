// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/cc_storage/internal/types"
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
		for denom, params := range defaultGenesis.CurrenciesParams {
			curInfo, err := keeper.GetResStdCurrencyInfo(ctx, denom)
			require.NoError(t, err)

			require.EqualValues(t, denom, curInfo.Denom)
			require.EqualValues(t, params.Decimals, curInfo.Decimals)
			require.EqualValues(t, common_vm.StdLibAddress, curInfo.Owner)
			require.EqualValues(t, 0, curInfo.TotalSupply.Uint64())
			require.False(t, curInfo.IsToken)

			curBalancePath, err := keeper.GetCurrencyBalancePath(ctx, denom)
			require.NoError(t, err)
			curInfoPath, err := keeper.GetCurrencyInfoPath(ctx, denom)
			require.NoError(t, err)

			require.EqualValues(t, params.BalancePath(), curBalancePath)
			require.EqualValues(t, params.InfoPath(), curInfoPath)
		}
	}

	// fail: non-existing denom
	{
		_, err := keeper.GetResStdCurrencyInfo(ctx, "nonexisting")
		require.Error(t, err)
	}
}
