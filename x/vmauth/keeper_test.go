// +build unit

package vmauth

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/common_vm"
)

// Test set/get account with empty balances.
func TestVMAccountKeeper_SetAccountEmptyBalances(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	balances := input.accountKeeper.loadBalances(input.ctx, addr)
	require.Empty(t, balances, "balances not empty")
	require.Empty(t, getter.GetCoins(), "coins not empty")
}

// Test set/get new account with balance.
func TestVMAccountKeeper_SetAccount(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	coin := types.NewCoin("dfi", types.NewInt(1))

	if err := acc.SetCoins(types.Coins{coin}); err != nil {
		t.Fatal(err)
	}

	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)

	balances := input.accountKeeper.loadBalances(input.ctx, addr)
	require.Len(t, balances, 1, "balances length not equal to 1 (saved one)")
	vmCoins := balancesToCoins(balances)
	require.True(t, getter.GetCoins().IsEqual(vmCoins), "coins are not match after set account")
}

// Test event handler generator creation if account not exists yet in VM storage.
func TestVMAccount_EventHandlerGeneratorNewAccount(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	// check if event generator created correctly.
	bz := input.vmStorage.GetValue(input.ctx, &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetEHPath(),
	})

	eventHandleGen := BytesToEventHandlerGen(bz)
	require.EqualValues(t, eventHandleGen.Counter, 2, "event handle generator has wrong counter")
	require.EqualValues(t, eventHandleGen.Addr, common_vm.Bech32ToLibra(addr))
}

// Test event handler generator when account already exists in VM storage.
func TestVMAccount_EventHandlerGenerator(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	// create new account in vm storage.
	vmAccount, eventHandleGen := CreateVMAccount(&acc)
	eventHandleGen.Counter = 10

	input.accountKeeper.saveNewVMAccount(input.ctx, addr, vmAccount, eventHandleGen)

	// now set account and see if event handle doesn't change.
	input.accountKeeper.SetAccount(input.ctx, &acc)
	bz := input.vmStorage.GetValue(input.ctx, &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetEHPath(),
	})

	loadedEventHandleGen := BytesToEventHandlerGen(bz)
	require.Equal(t, eventHandleGen.Counter, loadedEventHandleGen.Counter, "event handle generator has wrong counter")
	require.True(t, bytes.Equal(eventHandleGen.Addr, loadedEventHandleGen.Addr), "event handler addresses dont match")
}

// Test creation of event handlers for new account.
func TestVMAccount_EventHandlersNewAccount(t *testing.T) {
	// Test event handlers creation.
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	input.accountKeeper.SetAccount(input.ctx, &acc)

	// load account events
	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetResPath(),
	}

	bz := input.vmStorage.GetValue(input.ctx, accessPath)
	accRes := BytesToAccRes(bz)

	require.EqualValues(t, accRes.SentEvents.Counter, 0, "wrong counter for sent events")
	require.EqualValues(t, accRes.ReceivedEvents.Counter, 0, "wrong counter for received events")
	require.True(t, bytes.Equal(accRes.SentEvents.Guid, getGUID(addr, 0)), "wrong guid for sent events")
	require.True(t, bytes.Equal(accRes.ReceivedEvents.Guid, getGUID(addr, 1)), "wrong guid for received events")
}

// Test that already storead account in VM has still same event handlers.
func TestVMAccount_EventHandlers(t *testing.T) {
	// Test event handlers creation.
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	accRes, eventGen := CreateVMAccount(&acc)
	accRes.SentEvents.Counter = 10
	accRes.ReceivedEvents.Counter = 101

	input.accountKeeper.saveNewVMAccount(input.ctx, addr, accRes, eventGen)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	// load account events
	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetResPath(),
	}

	bz := input.vmStorage.GetValue(input.ctx, accessPath)
	loadedAccRes := BytesToAccRes(bz)

	require.Equal(t, accRes.SentEvents.Counter, loadedAccRes.SentEvents.Counter, "wrong counter for sent events")
	require.Equal(t, accRes.ReceivedEvents.Counter, loadedAccRes.ReceivedEvents.Counter, "wrong counter for received events")
	require.True(t, bytes.Equal(accRes.SentEvents.Guid, loadedAccRes.SentEvents.Guid), "wrong guid for sent events")
	require.True(t, bytes.Equal(accRes.ReceivedEvents.Guid, loadedAccRes.ReceivedEvents.Guid), "wrong guid for received events")
}

// Check get on account exists only in VM (check how balances works).
func TestVMAccount_GetExistsAccount(t *testing.T) {
	// Save account into.
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	// Set coins.
	dfiCoin := types.NewCoin("dfi", types.NewInt(100500))
	ethCoin := types.NewCoin("eth", types.NewInt(100200))
	coins := types.Coins{dfiCoin, ethCoin}

	if err := acc.SetCoins(coins); err != nil {
		t.Fatal(err)
	}

	balances, toDelete := coinsToBalances(&acc)

	require.Len(t, balances, len(coins), "balances length doesnt match coins")

	accRes, eventGen := CreateVMAccount(&acc)

	input.accountKeeper.saveNewVMAccount(input.ctx, addr, accRes, eventGen)
	input.accountKeeper.saveBalances(input.ctx, balances, toDelete)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	realCoins := balancesToCoins(balances)

	require.True(t, realCoins.IsEqual(getter.GetCoins()), "coins doesnt match")
}

// Save new virtual machine account.
func TestVMAccount_saveNewVMAccount(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	accRes := AccountResource{
		SentEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
		ReceivedEvents: &EventHandle{
			Counter: 1,
			Guid:    make([]byte, 40),
		},
	}

	eventHandleGen := EventHandleGenerator{
		Counter: 0,
		Addr:    common_vm.Bech32ToLibra(addr),
	}

	input.accountKeeper.saveNewVMAccount(input.ctx, addr, accRes, eventHandleGen)
	loadedAccRes, exists := input.accountKeeper.getVMAccount(input.ctx, addr)

	require.True(t, exists, "saved account doesnt exists")
	require.EqualValues(t, accRes, loadedAccRes, "saved and loaded accounts dont match")

	// load event generator
	bz := input.vmStorage.GetValue(input.ctx, &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetEHPath(),
	})

	loadedEventHandleGen := BytesToEventHandlerGen(bz)
	require.EqualValues(t, loadedEventHandleGen, eventHandleGen)
}

// Test set/get VM account.
func TestVMAccount_getVMAccount(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	accRes := AccountResource{
		SentEvents: &EventHandle{
			Counter: 0,
			Guid:    make([]byte, 40),
		},
		ReceivedEvents: &EventHandle{
			Counter: 1,
			Guid:    make([]byte, 40),
		},
	}

	input.accountKeeper.setVMAccount(input.ctx, addr, accRes)
	loadedAccRes, exists := input.accountKeeper.getVMAccount(input.ctx, addr)

	require.True(t, exists, "saved account doesnt exists")
	require.EqualValues(t, accRes, loadedAccRes, "saved and loaded accounts dont match")
}

// Test save/load balances.
func TestVMAccount_loadBalances(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()

	dfiCoin := types.NewCoin("dfi", types.NewInt(100500))
	ethCoin := types.NewCoin("eth", types.NewInt(100200))
	coins := types.Coins{dfiCoin, ethCoin}

	acc := auth.NewBaseAccountWithAddress(addr)
	if err := acc.SetCoins(coins); err != nil {
		t.Fatal(err)
	}

	balances, toDelete := coinsToBalances(&acc)

	input.accountKeeper.saveBalances(input.ctx, balances, toDelete)

	loadedBalances := input.accountKeeper.loadBalances(input.ctx, addr)
	realBalances := balancesToCoins(loadedBalances)

	require.True(t, coins.IsEqual(realBalances))
}

// Test remove account.
func TestVMAccountKeeper_RemoveAccount(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	input.accountKeeper.SetAccount(input.ctx, &acc)
	input.accountKeeper.RemoveAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.Nil(t, getter)

	key := &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(addr),
		Path:    GetResPath(),
	}

	bz := input.vmStorage.GetValue(input.ctx, key)
	require.Nil(t, bz)

	balances := loadAccessPaths(addr)
	for _, b := range balances {
		bz := input.vmStorage.GetValue(input.ctx, b.accessPath)
		if bz != nil {
			require.Nil(t, bz, "still contains coins balances after removing account")
		}
	}
}

// Test get signer acc.
func TestGetSignerAcc(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter, err := GetSignerAcc(input.ctx, input.accountKeeper, addr)
	require.EqualValues(t, &acc, getter)
	require.NoError(t, err)

	getter, err = GetSignerAcc(input.ctx, input.accountKeeper, types.AccAddress("bmp"))
	require.Nil(t, getter)
	require.Error(t, err)
}
