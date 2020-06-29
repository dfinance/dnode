// +build unit

package vmauth

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/common_vm"
)

const (
	balanceToDecode = "e8030000000000000000000000000000"
)

// Marshal balance.
func TestMarshalBalance(t *testing.T) {
	balance := BalanceResource{
		Value: types.NewInt(1000).BigInt(),
	}

	BalanceToBytes(balance)
}

// Unmarshal balance.
func TestUnmarshalBalance(t *testing.T) {
	bz, err := hex.DecodeString(balanceToDecode)
	require.NoError(t, err)

	balance := BytesToBalance(bz)
	require.Equal(t, balance.Value.String(), "1000")
}

// Coins to balances, and back.
func TestCoinsToBalances(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	coins := types.Coins{
		types.NewCoin("dfi", types.NewInt(1000)),
		types.NewCoin("eth", types.NewInt(100500)),
	}

	err := acc.SetCoins(coins)
	require.NoError(t, err)

	balances, _ := coinsToBalances(&acc)
	require.Len(t, balances, len(coins))

	for i := range coins {
		require.EqualValues(t, coins[i].Denom, balances[i].denom)
		require.EqualValues(t, coins[i].Amount.String(), balances[i].balance.Value.String())
	}
}

// Balances to coins, and back.
func TestBalancesToCoins(t *testing.T) {
	balances := Balances{
		Balance{
			denom: "dfi",
			balance: BalanceResource{
				Value: types.NewInt(10).BigInt(),
			},
		},
		Balance{
			denom: "eth",
			balance: BalanceResource{
				Value: types.NewInt(100500).BigInt(),
			},
		},
	}

	coins := balancesToCoins(balances)
	require.Len(t, coins, len(balances))

	for i := range balances {
		require.EqualValues(t, balances[i].denom, coins[i].Denom)
		require.EqualValues(t, balances[i].balance.Value.String(), coins[i].Amount.String())
	}
}

// Convert balances to bytes and back.
func TestBalanceToBytes(t *testing.T) {
	balances := Balances{
		Balance{
			denom: "dfi",
			balance: BalanceResource{
				Value: types.NewInt(10).BigInt(),
			},
		},
		Balance{
			denom: "eth",
			balance: BalanceResource{
				Value: types.NewInt(100500).BigInt(),
			},
		},
	}

	for _, balance := range balances {
		bz := BalanceToBytes(balance.balance)
		loadedBalance := BytesToBalance(bz)

		require.EqualValues(t, balance.balance, loadedBalance)
	}
}

// Test access paths loading.
func TestLoadAccessPaths(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	balances := loadAccessPaths(addr)

	for _, balance := range balances {
		denomPath, isOk := denomPaths[balance.denom]

		require.True(t, isOk)
		require.EqualValues(t, denomPaths[balance.denom], denomPath)

		acсessPath := &vm_grpc.VMAccessPath{
			Address: common_vm.Bech32ToLibra(addr),
			Path:    denomPath,
		}

		require.EqualValues(t, acсessPath, balance.accessPath)
	}
}
