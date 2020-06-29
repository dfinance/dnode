// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
