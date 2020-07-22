// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
)

// Test NewCurrencyInfo.
func TestCCS_NewCurrencyInfo(t *testing.T) {
	t.Parallel()

	currency := NewCurrency(
		CurrencyParams{
			Denom:          "test",
			Decimals:       4,
			BalancePathHex: "",
			InfoPathHex:    "",
		},
		sdk.NewIntFromUint64(100),
	)

	// ok: stdlib
	{
		curInfo, err := NewResCurrencyInfo(currency, common_vm.StdLibAddress)
		require.NoError(t, err)
		require.EqualValues(t, currency.Denom, curInfo.Denom)
		require.EqualValues(t, currency.Decimals, curInfo.Decimals)
		require.EqualValues(t, currency.Supply.Uint64(), curInfo.TotalSupply.Uint64())
		require.EqualValues(t, common_vm.StdLibAddress, curInfo.Owner)
		require.False(t, curInfo.IsToken)
	}

	// ok: token
	{
		owner := make([]byte, common_vm.VMAddressLength)

		curInfo, err := NewResCurrencyInfo(currency, owner)
		require.NoError(t, err)
		require.EqualValues(t, currency.Denom, curInfo.Denom)
		require.EqualValues(t, currency.Decimals, curInfo.Decimals)
		require.EqualValues(t, currency.Supply.Uint64(), curInfo.TotalSupply.Uint64())
		require.EqualValues(t, owner, curInfo.Owner)
		require.True(t, curInfo.IsToken)
	}

	// fail
	{
		owner := make([]byte, common_vm.VMAddressLength-1)

		_, err := NewResCurrencyInfo(currency, owner)
		require.Error(t, err)
	}
}
