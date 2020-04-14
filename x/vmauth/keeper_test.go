// +build unit

package vmauth

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
)

func TestVMAccountKeeper_SetAccount(t *testing.T) {
	input := newTestInput(t)

	// Check just set.
	addr := types.AccAddress("tmp1")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	// Check set with coins.
	coin := types.NewCoin("dfi", types.NewInt(1))
	acc = auth.NewBaseAccountWithAddress(addr)
	if err := acc.SetCoins(types.Coins{coin}); err != nil {
		t.Fatal(err)
	}
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter = input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)
}

func TestVMAccountKeeper_SetAccountEventHandler(t *testing.T) {
	input := newTestInput(t)
	addr := types.AccAddress("tmp1")

	coin := types.NewCoin("dfi", types.NewInt(1))
	acc := auth.NewBaseAccountWithAddress(addr)
	if err := acc.SetCoins(types.Coins{coin}); err != nil {
		t.Fatal(err)
	}

	// prepare bz
	accRes := AccountResource{
		WithdrawEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		DepositEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		EventGenerator: 2,
	}

	path := &vm_grpc.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	}

	input.vmStorage.SetValue(input.ctx, path, AccResToBytes(accRes))
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	bz := input.vmStorage.GetValue(input.ctx, path)
	require.NotNil(t, bz)

	accRes2 := BytesToAccRes(bz)
	for i, coin := range getter.GetCoins() {
		require.EqualValues(t, coin.Denom, accRes2.Balances[i].Denom)
		require.EqualValues(t, coin.Amount.String(), accRes2.Balances[i].Value.String())
	}

	require.Equal(t, accRes.EventGenerator, accRes2.EventGenerator)
	require.EqualValues(t, accRes.WithdrawEvents, accRes2.WithdrawEvents)
	require.EqualValues(t, accRes.DepositEvents, accRes2.DepositEvents)
}

func TestVMAccountKeeper_GetAccount(t *testing.T) {
	input := newTestInput(t)

	// Check just get with vm storage.
	coin := types.NewCoin("dfi", types.NewInt(1))

	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	if err := acc.SetCoins(types.Coins{coin}); err != nil {
		t.Fatal(err)
	}
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	key := &vm_grpc.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	}

	accData := input.vmStorage.GetValue(input.ctx, key)
	accRes := BytesToAccRes(accData)

	require.Len(t, acc.GetCoins(), len(accRes.Balances))

	accData = input.vmStorage.GetValue(input.ctx, key)
	accRes = BytesToAccRes(accData)

	require.Len(t, acc.Coins, len(accRes.Balances))

	for i, coin := range getter.GetCoins() {
		require.EqualValues(t, coin.Denom, accRes.Balances[i].Denom)
		require.EqualValues(t, coin.Amount.String(), accRes.Balances[i].Value.String())
	}

	// Check get if there is write in vm storage, but not in cosmos storage.
	vmAcc := auth.NewBaseAccountWithAddress(types.AccAddress("vm"))
	if err := vmAcc.SetCoins(types.Coins{coin}); err != nil {
		t.Fatal(err)
	}
	vmKey := &vm_grpc.VMAccessPath{
		Address: AddrToPathAddr(vmAcc.GetAddress()),
		Path:    GetResPath(),
	}

	bz := AccResToBytes(AccResFromAccount(&vmAcc, nil))

	input.vmStorage.SetValue(input.ctx, vmKey, bz)

	getter = input.accountKeeper.GetAccount(input.ctx, vmAcc.GetAddress())
	require.EqualValues(t, &vmAcc, getter)

	// Returns nil.
	getter = input.accountKeeper.GetAccount(input.ctx, types.AccAddress("nil"))
	require.Nil(t, getter)
}

func TestVMAccountKeeper_RemoveAccount(t *testing.T) {
	input := newTestInput(t)
	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	input.accountKeeper.RemoveAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.Nil(t, getter)

	key := &vm_grpc.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	}

	bz := input.vmStorage.GetValue(input.ctx, key)
	require.Empty(t, bz)
}

func TestGetSignerAcc(t *testing.T) {
	input := newTestInput(t)
	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter, err := GetSignerAcc(input.ctx, input.accountKeeper, addr)
	require.EqualValues(t, &acc, getter)
	require.NoError(t, err)

	getter, err = GetSignerAcc(input.ctx, input.accountKeeper, types.AccAddress("bmp"))
	require.Nil(t, getter)
	require.Error(t, err)
}
