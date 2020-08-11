// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/glav"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// Test keeper CreateCurrency method.
func TestCCSKeeper_CreateCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	params := types.CurrencyParams{
		Denom:    "test",
		Decimals: 8,
	}
	denom := params.Denom

	// ok
	{
		err := keeper.CreateCurrency(ctx, params)
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
		curBalancePath := glav.BalanceVector(denom)
		curInfoPath := glav.CurrencyInfoVector(denom)

		require.EqualValues(t, currency.BalancePath(), curBalancePath)
		require.EqualValues(t, currency.InfoPath(), curInfoPath)
	}

	// fail: existing
	{
		err := keeper.CreateCurrency(ctx, params)
		require.Error(t, err)
	}
}

// Test keeper GetCurrency method.
func TestCCSKeeper_GetCurrency(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	// create currency
	params := types.CurrencyParams{
		Denom:    "test",
		Decimals: uint8(8),
	}
	denom := params.Denom

	err := keeper.CreateCurrency(ctx, params)
	require.NoError(t, err)

	// ok
	{
		currency, err := keeper.GetCurrency(ctx, denom)
		require.NoError(t, err)

		require.Equal(t, denom, currency.Denom)
		require.EqualValues(t, params.Decimals, currency.Decimals)
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

// Test keeper IncreaseCurrencySupply / DecreaseCurrencySupply methods.
func TestCCSKeeper_Supply(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	// create currency
	curAmount := sdk.ZeroInt()
	params := types.CurrencyParams{
		Denom:    "test",
		Decimals: uint8(8),
	}
	denom := params.Denom

	err := keeper.CreateCurrency(ctx, params)
	require.NoError(t, err)

	// ok: increase
	{
		amount := sdk.NewIntFromUint64(100)
		coin := sdk.NewCoin(denom, amount)
		curAmount = curAmount.Add(amount)

		err := keeper.IncreaseCurrencySupply(ctx, coin)
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

		coin := sdk.NewCoin(denom, amount)
		err := keeper.DecreaseCurrencySupply(ctx, coin)
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
		coin := sdk.NewCoin("invalid", sdk.OneInt())

		require.Error(t, keeper.IncreaseCurrencySupply(ctx, coin))
		require.Error(t, keeper.DecreaseCurrencySupply(ctx, coin))
	}
}

func TestCCSKeeper_GetCurrencies(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	defGenesis := types.DefaultGenesisState()

	currencies := keeper.GetCurrencies(ctx)
	require.Len(t, currencies, len(defGenesis.CurrenciesParams))
}
