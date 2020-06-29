// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Test keeper DestroyCurrency method.
func TestCurrenciesKeeper_DestroyCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")

	// fail: unknown currency
	{
		require.Error(t, keeper.DestroyCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// ok
	{
		destroyID := keeper.getNextDestroyID(ctx)
		require.False(t, keeper.HasDestroy(ctx, destroyID))
		require.NoError(t, keeper.DestroyCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))
		require.True(t, keeper.HasDestroy(ctx, destroyID))

		require.True(t, input.bankKeeper.GetCoins(ctx, addr).AmountOf(defDenom).IsZero())

		currency, err := keeper.GetCurrency(ctx, defDenom)
		require.NoError(t, err)
		require.True(t, currency.Supply.IsZero())
	}

	// fail: insufficient coins (balance is 0)
	{
		require.Error(t, keeper.DestroyCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// fail: insufficient coins (account doesn't have denom currency)
	{
		require.Error(t, keeper.DestroyCurrency(ctx, "otherdenom", defAmount, addr, recipient.String(), ctx.ChainID()))
	}
}

// Test keeper GetDestroy method.
func TestCurrenciesKeeper_GetDestroy(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, defAmount, defDecimals, addr))

	// destroy currency
	require.NoError(t, keeper.DestroyCurrency(ctx, defDenom, defAmount, addr, recipient.String(), ctx.ChainID()))

	// ok
	{
		id := keeper.getLastDestroyID(ctx)
		destroy, err := keeper.GetDestroy(ctx, id)
		require.NoError(t, err)

		require.Equal(t, id.String(), destroy.ID.String())
		require.Equal(t, defDenom, destroy.Denom)
		require.True(t, destroy.Amount.Equal(defAmount))
		require.Equal(t, addr.String(), destroy.Spender.String())
		require.Equal(t, recipient.String(), destroy.Recipient)
		require.Equal(t, ctx.ChainID(), destroy.ChainID)
		require.True(t, keeper.HasDestroy(ctx, id))
	}

	// fail: non-existing
	{
		_, err := keeper.GetDestroy(ctx, keeper.getNextDestroyID(ctx))
		require.Error(t, err)
	}
}

// Test keeper GetDestroysFiltered method.
func TestCurrenciesKeeper_GetDestroysFiltered(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	addr := input.CreateAccount(t, "addr1", nil)
	ctx, keeper := input.ctx, input.keeper

	recipient := sdk.AccAddress("addr2")
	destroysCount := 5
	amount := sdk.NewIntFromUint64(100)
	destroyAmount := amount.QuoRaw(int64(destroysCount))

	// issue currency
	require.NoError(t, keeper.IssueCurrency(ctx, defIssueID1, defDenom, amount, defDecimals, addr))

	// multiple destroys
	for i := 0; i < destroysCount; i++ {
		require.NoError(t, keeper.DestroyCurrency(ctx, defDenom, destroyAmount, addr, recipient.String(), ctx.ChainID()))
	}

	// request all
	{
		params := types.DestroysReq{
			Page:  sdk.NewUint(1),
			Limit: sdk.NewUint(10),
		}
		destroys, err := keeper.GetDestroysFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, destroys, destroysCount)

		for i, destroy := range destroys {
			id := dnTypes.NewIDFromUint64(uint64(i))
			require.True(t, destroy.ID.Equal(id))
		}
	}

	// request page 1
	{
		params := types.DestroysReq{
			Page:  sdk.NewUint(1),
			Limit: sdk.NewUint(3),
		}
		destroys, err := keeper.GetDestroysFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, destroys, 3)

		for i, destroy := range destroys {
			id := dnTypes.NewIDFromUint64(uint64(i))
			require.True(t, destroy.ID.Equal(id))
		}
	}

	// request page 2
	{
		params := types.DestroysReq{
			Page:  sdk.NewUint(2),
			Limit: sdk.NewUint(3),
		}
		destroys, err := keeper.GetDestroysFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, destroys, 2)

		for i, destroy := range destroys {
			id := dnTypes.NewIDFromUint64(uint64(3 + i))
			require.True(t, destroy.ID.Equal(id))
		}
	}

	// request wrong page
	{
		params := types.DestroysReq{
			Page:  sdk.NewUint(3),
			Limit: sdk.NewUint(3),
		}
		destroys, err := keeper.GetDestroysFiltered(ctx, params)
		require.NoError(t, err)
		require.Len(t, destroys, 0)
	}
}
