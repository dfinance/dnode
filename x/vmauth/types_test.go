// +build unit

package vmauth

import (
	"encoding/binary"
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
	accToDecode     = "0000000000000000280000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002800000000000000000000000000000000000000000000000000000000000000000000000000000000"
	balanceToDecode = "e8030000000000000000000000000000"
)

var (
	eventHandlerGenToDecode []byte
	eventHandlerGenAcc      types.AccAddress
)

// Marshal account.
func TestMarshalAccount(t *testing.T) {
	accRes := AccountResource{
		SentEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
		ReceivedEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
	}
	AccResToBytes(accRes)
}

// Unmarshal account.
func TestUnmarshalAccount(t *testing.T) {
	accRes := AccountResource{
		SentEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
		ReceivedEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
	}

	bz, err := hex.DecodeString(accToDecode)
	require.NoError(t, err)

	loaded := BytesToAccRes(bz)
	require.EqualValues(t, accRes, loaded)
}

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

// Marshal event handler gen.
func TestEventHandlerGenMarshal(t *testing.T) {
	eventHandlerGenAcc = secp256k1.GenPrivKey().PubKey().Address().Bytes()

	eventHandleGen := EventHandleGenerator{
		Counter: 0,
		Addr:    common_vm.Bech32ToLibra(eventHandlerGenAcc),
	}

	eventHandlerGenToDecode = EventHandlerGenToBytes(eventHandleGen)
}

// Unmarshal event handler gen.
func TestEventHandlerGenUnmarshal(t *testing.T) {
	eventHandlerGen := BytesToEventHandlerGen(eventHandlerGenToDecode)
	require.EqualValues(t, eventHandlerGen.Counter, 0)
	require.EqualValues(t, eventHandlerGen.Addr, common_vm.Bech32ToLibra(eventHandlerGenAcc))
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

// Testing get GUID.
func TestGetGUID(t *testing.T) {
	var counter uint64 = 0
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	guid := getGUID(addr, 0)

	countBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(countBytes, counter)
	realGuid := append(countBytes, common_vm.Bech32ToLibra(addr)...)

	require.EqualValues(t, realGuid, guid)
}

// Create VM account test.
func TestCreateVMAccount(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	vmAcc, eventHandlerGenerator := CreateVMAccount(&acc)

	require.EqualValues(t, getGUID(addr, 0), vmAcc.SentEvents.Guid)
	require.EqualValues(t, getGUID(addr, 1), vmAcc.ReceivedEvents.Guid)
	require.EqualValues(t, 0, vmAcc.SentEvents.Counter)
	require.EqualValues(t, 0, vmAcc.ReceivedEvents.Counter)
	require.EqualValues(t, 2, eventHandlerGenerator.Counter)
	require.EqualValues(t, common_vm.Bech32ToLibra(addr), eventHandlerGenerator.Addr)
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

// Test get rest path.
func TestGetResPath(t *testing.T) {
	path := GetResPath()
	bz, err := hex.DecodeString(resourceKey)
	require.NoError(t, err)
	require.EqualValues(t, path, bz)
}

// Test get event handler generator path.
func TestGetEHPath(t *testing.T) {
	path := GetEHPath()
	bz, err := hex.DecodeString(ehResourceKey)
	require.NoError(t, err)
	require.EqualValues(t, path, bz)
}
