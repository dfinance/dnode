// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Test keeper CreateCurrency method.
func TestCurrenciesKeeper_CreateCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	denom := "test"
	params := types.CurrencyParams{
		Decimals:       8,
		BalancePathHex: "010203",
		InfoPathHex:    "AABBCC",
	}

	// ok
	{
		err := keeper.CreateCurrency(ctx, denom, params)
		require.NoError(t, err)

		// check currency
		currency, err := keeper.GetCurrency(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, denom, currency.Denom)
		require.EqualValues(t, params.Decimals, currency.Decimals)
		require.True(t, currency.Supply.IsZero())
		require.True(t, keeper.HasCurrency(ctx, denom))

		// check currencyInfo
		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, denom)
		require.NoError(t, err)
		require.EqualValues(t, denom, curInfo.Denom)
		require.EqualValues(t, params.Decimals, curInfo.Decimals)
		require.EqualValues(t, common_vm.StdLibAddress, curInfo.Owner)
		require.EqualValues(t, 0, curInfo.TotalSupply.Uint64())
		require.False(t, curInfo.IsToken)

		// check VM paths
		curBalancePath, err := keeper.GetCurrencyBalancePath(ctx, denom)
		require.NoError(t, err)
		curInfoPath, err := keeper.GetCurrencyInfoPath(ctx, denom)
		require.NoError(t, err)

		require.EqualValues(t, params.BalancePath(), curBalancePath)
		require.EqualValues(t, params.InfoPath(), curInfoPath)
	}

	// fail: existing
	{
		err := keeper.CreateCurrency(ctx, denom, params)
		require.Error(t, err)
	}
}

// Test keeper GetCurrency method.
func TestCurrenciesKeeper_GetCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// ok
	{
		currency, err := keeper.GetCurrency(ctx, defDenom)
		require.NoError(t, err)

		require.Equal(t, defDenom, currency.Denom)
		require.EqualValues(t, defDecimals, currency.Decimals)
		require.True(t, currency.Supply.Equal(defAmount))
		require.True(t, keeper.HasCurrency(ctx, defDenom))
	}

	// fail: non-existing
	{
		nonExistingDenom := "wrongdenom"
		_, err := keeper.GetCurrency(ctx, nonExistingDenom)
		require.Error(t, err)
		require.False(t, keeper.HasCurrency(ctx, nonExistingDenom))
	}
}
