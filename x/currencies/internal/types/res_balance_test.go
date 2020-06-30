// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
)

// Test checks balance resource lcs marshalling.
func TestCurrencies_ResBalance(t *testing.T) {
	inRes := ResBalance{Value: sdk.NewIntFromUint64(1234567890).BigInt()}

	inBz, err := inRes.Bytes()
	require.NoError(t, err)

	outRes, err := NewResBalance(inBz)
	require.NoError(t, err)

	require.EqualValues(t, inRes.Value.Uint64(), outRes.Value.Uint64())
}

// Test check balance creation.
func TestCurrencies_Balance(t *testing.T) {
	res := ResBalance{Value: sdk.NewIntFromUint64(1234567890).BigInt()}
	resBz, err := res.Bytes()
	require.NoError(t, err)

	// ok
	{
		denom := "test"
		accessPath := &vm_grpc.VMAccessPath{
			Address: []byte{1, 2, 3},
			Path:    []byte{4, 5, 6},
		}
		balance, err := NewBalance(denom, accessPath, resBz)
		require.NoError(t, err)

		require.Equal(t, denom, balance.Denom)
		require.EqualValues(t, accessPath.Address, balance.AccessPath.Address)
		require.EqualValues(t, accessPath.Path, balance.AccessPath.Path)
		require.EqualValues(t, res.Value.Uint64(), balance.Resource.Value.Uint64())

		require.Equal(t, denom, balance.Coin().Denom)
		require.Equal(t, res.Value.String(), balance.Coin().Amount.String())
	}

	// fail: empty denom
	{
		_, err := NewBalance("", &vm_grpc.VMAccessPath{}, resBz)
		require.Error(t, err)
	}

	// fail: nil accessPath
	{
		_, err := NewBalance("test", nil, resBz)
		require.Error(t, err)
	}

	// fail: invalid resBz
	{
		_, err := NewBalance("test", &vm_grpc.VMAccessPath{}, nil)
		require.Error(t, err)
	}
}

// Test check balances function.
func TestCurrencies_Balances(t *testing.T) {
	balances := Balances{
		Balance{
			Denom:      "testa",
			AccessPath: nil,
			Resource:   ResBalance{ Value: sdk.NewIntFromUint64(1).BigInt()},
		},
		Balance{
			Denom:      "testb",
			AccessPath: nil,
			Resource:   ResBalance{ Value: sdk.NewIntFromUint64(2).BigInt()},
		},
	}

	coins := balances.Coins()
	require.Len(t, coins, len(balances))

	for i, balance := range balances {
		coin := coins[i]
		require.Equal(t, balance.Denom, coin.Denom)
		require.Equal(t, balance.Resource.Value, coin.Amount.BigInt())
	}
}
