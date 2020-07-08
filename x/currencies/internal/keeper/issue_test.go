// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test keeper IssueCurrency method.
func TestCurrenciesKeeper_IssueCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	// ok
	{
		require.False(t, keeper.HasIssue(ctx, defIssueID1))
		require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defCoin, addr))

		// check account balance changed
		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(defAmount))

		// check currency supply increased
		currency, err := input.ccsStorage.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.Equal(defAmount))

		// check currencyInfo supply increased
		curInfo, err := input.ccsStorage.GetResStdCurrencyInfo(ctx, defDenom)
		require.NoError(t, err)
		require.Equal(t, curInfo.TotalSupply.String(), defAmount.String())

		// check supply mod supply increased
		supply := input.supplyKeeper.GetSupply(ctx)
		for _, coin := range supply.GetTotal() {
			if coin.Denom == defDenom {
				require.Equal(t, coin.Amount.String(), defAmount.String())
			}
		}
	}

	// fail: existing issueID
	{
		require.Error(t, keeper.IssueCurrency(ctx, defIssueID1, defCoin, addr))
	}

	// ok: issue existing currency, increasing supply
	{
		newAmount := defAmount.MulRaw(2)

		require.False(t, keeper.HasIssue(ctx, defIssueID2))
		require.NoError(t, keeper.IssueCurrency(ctx, defIssueID2, defCoin, addr))

		// check account balance changed
		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(newAmount))

		// check currency supply increased
		currency, err := input.ccsStorage.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.Equal(newAmount))

		// check currencyInfo supply increased
		curInfo, err := input.ccsStorage.GetResStdCurrencyInfo(ctx, defDenom)
		require.NoError(t, err)
		require.Equal(t, curInfo.TotalSupply.String(), newAmount.String())
	}
}

// Test keeper GetIssue method.
func TestCurrenciesKeeper_GetIssue(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defCoin, addr))

	// ok
	{
		issue, err := keeper.GetIssue(ctx, defIssueID1)
		require.NoError(t, err)
		require.True(t, defCoin.IsEqual(issue.Coin))
		require.Equal(t, addr.String(), issue.Payee.String())
		require.True(t, keeper.HasIssue(ctx, defIssueID1))
	}

	// fail: non-existing
	{
		_, err := keeper.GetIssue(ctx, defIssueID2)
		require.Error(t, err)
		require.False(t, keeper.HasIssue(ctx, defIssueID2))
	}
}

// Test keeper IssueCurrency method: huge amount.
func TestCurrenciesKeeper_IssueHugeAmount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	amount, ok := sdk.NewIntFromString("1000000000000000000000000000000000000000000000")
	require.True(t, ok)
	coin := sdk.NewCoin(defDenom, amount)

	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, coin, addr))
	require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).Equal(amount))
}
