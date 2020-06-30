// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Test keeper WithdrawCurrency method.
func TestCurrenciesKeeper_WithdrawCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")

	// fail: unknown currency
	{
		require.Error(t, keeper.WithdrawCurrency(ctx, "test", defAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// ok
	{
		withdrawID := keeper.getNextWithdrawID(ctx)
		require.False(t, keeper.HasWithdraw(ctx, withdrawID))
		require.NoError(t, keeper.WithdrawCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))
		require.True(t, keeper.HasWithdraw(ctx, withdrawID))

		// check account balance changed
		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).IsZero())

		// check currency supply decreased
		currency, err := keeper.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.IsZero())

		// check currencyInfo supply decreased
		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, defDenom)
		require.NoError(t, err)
		require.Equal(t, curInfo.TotalSupply.String(), "0")
	}

	// fail: insufficient coins (balance is 0)
	{
		require.Error(t, keeper.WithdrawCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// fail: insufficient coins (account doesn't have denom currency)
	{
		require.Error(t, keeper.WithdrawCurrency(ctx, "otherdenom", defAmount, addr, recipient.String(), ctx.ChainID()))
	}
}

// Test keeper GetWithdraw method.
func TestCurrenciesKeeper_GetWithdraw(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// withdraw currency
	require.NoError(t, keeper.WithdrawCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))

	// ok
	{
		id := keeper.getLastWithdrawID(ctx)
		withdraw, err := keeper.GetWithdraw(ctx, id)
		require.NoError(t, err)

		require.Equal(t, id.String(), withdraw.ID.String())
		require.Equal(t, defDenom, withdraw.Denom)
		require.True(t, withdraw.Amount.Equal(defAmount))
		require.Equal(t, addr.String(), withdraw.Spender.String())
		require.Equal(t, recipient.String(), withdraw.PegZoneSpender)
		require.Equal(t, ctx.ChainID(), withdraw.PegZoneChainID)
		require.True(t, keeper.HasWithdraw(ctx, id))
	}

	// fail: non-existing
	{
		_, err := keeper.GetWithdraw(ctx, keeper.getNextWithdrawID(ctx))
		require.Error(t, err)
	}
}

// Test keeper GetWithdrawsFiltered method.
func TestCurrenciesKeeper_GetWithdrawsFiltered(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")
	withdrawCount := 5
	amount := sdk.NewIntFromUint64(100)
	withdrawAmount := amount.QuoRaw(int64(withdrawCount))

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, amount, defDecimals, addr))

	// multiple withdraws
	for i := 0; i < withdrawCount; i++ {
		require.NoError(t, keeper.WithdrawCurrency(ctx, defDenom, withdrawAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// request all
	{
		params := types.WithdrawsReq{
			Page:  sdk.NewUint(1),
			Limit: sdk.NewUint(10),
		}
		withdraws, err := keeper.GetWithdrawsFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, withdraws, withdrawCount)

		for i, withdraw := range withdraws {
			id := dnTypes.NewIDFromUint64(uint64(i))
			require.True(t, withdraw.ID.Equal(id))
		}
	}

	// request page 1
	{
		params := types.WithdrawsReq{
			Page:  sdk.NewUint(1),
			Limit: sdk.NewUint(3),
		}
		withdraws, err := keeper.GetWithdrawsFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, withdraws, 3)

		for i, withdraw := range withdraws {
			id := dnTypes.NewIDFromUint64(uint64(i))
			require.True(t, withdraw.ID.Equal(id))
		}
	}

	// request page 2
	{
		params := types.WithdrawsReq{
			Page:  sdk.NewUint(2),
			Limit: sdk.NewUint(3),
		}
		withdraws, err := keeper.GetWithdrawsFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, withdraws, 2)

		for i, withdraw := range withdraws {
			id := dnTypes.NewIDFromUint64(uint64(3 + i))
			require.True(t, withdraw.ID.Equal(id))
		}
	}

	// request wrong page
	{
		params := types.WithdrawsReq{
			Page:  sdk.NewUint(3),
			Limit: sdk.NewUint(3),
		}
		withdraws, err := keeper.GetWithdrawsFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, withdraws, 0)
	}
}
