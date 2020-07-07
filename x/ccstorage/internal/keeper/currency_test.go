// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// Test keeper CreateCurrency method.
func TestCCSKeeper_CreateCurrency(t *testing.T) {
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
func TestCCSKeeper_GetCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	// create currency
	denom, decimals := "test", uint8(8)
	params := types.CurrencyParams{
		Decimals:       decimals,
		BalancePathHex: "010203",
		InfoPathHex:    "AABBCC",
	}
	err := keeper.CreateCurrency(ctx, denom, params)
	require.NoError(t, err)

	// ok
	{
		currency, err := keeper.GetCurrency(ctx, denom)
		require.NoError(t, err)

		require.Equal(t, denom, currency.Denom)
		require.EqualValues(t, decimals, currency.Decimals)
		require.True(t, currency.Supply.IsZero())
		require.True(t, keeper.HasCurrency(ctx, denom))
	}

	// fail: non-existing
	{
		nonExistingDenom := "wrongdenom"
		_, err := keeper.GetCurrency(ctx, nonExistingDenom)
		require.Error(t, err)
		require.False(t, keeper.HasCurrency(ctx, nonExistingDenom))
	}
}

func TestCCSKeeper_Supply(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	// create currency
	denom, decimals, curAmount := "test", uint8(8), sdk.ZeroInt()
	params := types.CurrencyParams{
		Decimals:       decimals,
		BalancePathHex: "010203",
		InfoPathHex:    "AABBCC",
	}
	err := keeper.CreateCurrency(ctx, denom, params)
	require.NoError(t, err)

	// ok: increase
	{
		amount := sdk.NewIntFromUint64(100)
		curAmount = curAmount.Add(amount)

		err := keeper.IncreaseCurrencySupply(ctx, denom, amount)
		require.NoError(t, err)

		currency, err := keeper.GetCurrency(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, curAmount.String(), currency.Supply.String())

		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, curAmount.String(), curInfo.TotalSupply.String())
	}

	// ok: decrease
	{
		amount := sdk.NewIntFromUint64(50)
		curAmount = curAmount.Sub(amount)

		err := keeper.DecreaseCurrencySupply(ctx, denom, amount)
		require.NoError(t, err)

		currency, err := keeper.GetCurrency(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, curAmount.String(), currency.Supply.String())

		curInfo, err := keeper.GetResStdCurrencyInfo(ctx, denom)
		require.NoError(t, err)
		require.Equal(t, curAmount.String(), curInfo.TotalSupply.String())
	}

	// fail: non-existing currency
	{
		require.Error(t, keeper.IncreaseCurrencySupply(ctx, "invalid", sdk.OneInt()))

		require.Error(t, keeper.DecreaseCurrencySupply(ctx, "invalid", sdk.OneInt()))
	}
}
